package dnsserver

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type (
	DNSRegisterer interface {
		Register(domain string, ip string)
	}
	DNSUnRegisterer interface {
		Unregister(domain string)
	}
	AliasFileLoader interface {
		LoadAliasesFromFile()
	}
	IPResolver interface {
		GetIPFromDomain(string) (string, bool)
	}
	DNSRegistry interface {
		DNSRegisterer
		DNSUnRegisterer
		AliasFileLoader
		IPResolver
	}
)

type DNSUpdater struct {
}

func NewDNSUpdater() *DNSUpdater {
	return &DNSUpdater{}
}

func (u *DNSUpdater) Start(ctx context.Context, dnsRegistry DNSRegistry) {
	go func() {
		u.startEventListener(ctx, dnsRegistry)
	}()
}

func (u *DNSUpdater) startEventListener(ctx context.Context, dnsRegistry DNSRegistry) {
	dockerclient, dockerClientDefer := u.getDockerClient()
	defer dockerClientDefer()

	eventFilter := filters.NewArgs()
	eventFilter.Add("type", "container")
	options := types.EventsOptions{
		Filters: eventFilter,
	}
	evtCh, errCh := dockerclient.Events(ctx, options)

	for {
		select {
		case <-errCh:
			fmt.Println("error in docker event loop, stopping Docker DNS Survey")
			return
		case e := <-evtCh:
			switch e.Action {
			case "kill", "die", "stop":
				u.removeContainerFromDNS(dockerclient, e, dnsRegistry)
			case "start":
				u.addContainerToDNS(dockerclient, e, dnsRegistry)
			}
		case <-ctx.Done():
			fmt.Println("Stopping Docker DNS Survey")
			return
		}
	}
}

func (u *DNSUpdater) addContainerToDNS(dockerclient *client.Client, e events.Message, dnsRegistry DNSRegistry) {
	ip, err := u.getContainerIp(dockerclient, e.Actor.ID)
	if err != nil {
		logrus.Errorf("could not determine container ip: %v", err)
		return
	}
	containerName, err := u.getContainerName(dockerclient, e.Actor.ID)
	if err != nil {
		logrus.Errorf("could not determine container name: %v", err)
		return
	}
	logrus.Infof("adding container %s due to (%s) event", containerName, e.Action)
	dnsRegistry.Register(containerName, ip)
}

func (u *DNSUpdater) removeContainerFromDNS(dockerclient *client.Client, e events.Message, dnsRegistry DNSRegistry) {
	containerName, err := u.getContainerName(dockerclient, e.Actor.ID)
	if err != nil {
		logrus.Errorf("could not determine container name: %v", err)
		return
	}
	logrus.Infof("removing container %s due to (%s) event", containerName, e.Action)

	dnsRegistry.Unregister(containerName)
}

func (u *DNSUpdater) getContainerName(dockerclient *client.Client, containerID string) (string, error) {
	container, err := u.getContainerByID(dockerclient, containerID)
	if err != nil {
		return "", fmt.Errorf("error getting container name: %v", err)
	}

	return container.Names[0], nil
}

func (u *DNSUpdater) getContainerIp(dockerclient *client.Client, containerID string) (string, error) {
	container, err := u.getContainerByID(dockerclient, containerID)
	if err != nil {
		return "", fmt.Errorf("error getting container IP: %v", err)
	}

	ips := GetContainerNetworkIps(container)

	return ips[len(ips)-1], nil
}

func (u *DNSUpdater) getContainerByID(dockerclient *client.Client, containerID string) (types.Container, error) {
	ctx := context.Background()
	containerFilter := filters.NewArgs()
	containerFilter.Add("id", containerID)
	options := types.ContainerListOptions{
		Filters: containerFilter,
		All:     true,
	}
	containers, err := dockerclient.ContainerList(ctx, options)
	if err != nil {
		return types.Container{}, fmt.Errorf("error while getting container by ID using container list: %v", err)
	}

	if len(containers) == 0 {
		return types.Container{}, fmt.Errorf("error while getting container by ID: no container found for id '%s'", containerID)
	}

	return containers[0], nil
}

func (u *DNSUpdater) getDockerClient() (*client.Client, func()) {
	dockerClient, err := client.NewEnvClient()

	return dockerClient, func() {
		err = dockerClient.Close()
		if err != nil {
			fmt.Printf("error closing docker dockerClient: %v", err)
		}
	}
}
