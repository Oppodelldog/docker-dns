package dnsserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

var ErrGettingContainerName = errors.New("error getting container name")
var ErrGettingContainerIP = errors.New("error getting container IP")
var ErrGettingContainerByID = errors.New("error while getting container by ID")
var ErrNoContainerFoundForID = errors.New("no container found")

type DNSUpdater struct {
	dockerClientAdapter DockerClientAdapter
	dockerClient        *client.Client
	ctx                 context.Context
	dnsRegistry         DNSRegistrar
}

func NewDNSUpdater(ctx context.Context,
	dockerClient *client.Client,
	dockerClientAdapter DockerClientAdapter,
	dnsRegistry DNSRegistrar,
) DNSUpdater {
	u := DNSUpdater{
		dockerClientAdapter: dockerClientAdapter,
		dockerClient:        dockerClient,
		ctx:                 ctx,
		dnsRegistry:         dnsRegistry,
	}

	u.start()

	return u
}

func (u DNSUpdater) start() {
	go func() {
		u.startEventListener()
	}()
}

func (u DNSUpdater) startEventListener() {
	evtCh, errCh := u.registerContainerEvents()

	for {
		select {
		case <-errCh:
			logrus.Errorf("error in docker event loop, stopping Docker DNS Survey")

			return
		case e := <-evtCh:
			switch e.Action {
			case "kill", "die", "stop":
				u.removeContainerFromDNS(e)
			case "start":
				u.addContainerToDNS(e)
			}
		case <-u.ctx.Done():
			logrus.Info("Stopping Docker DNS Survey")

			return
		}
	}
}

func (u DNSUpdater) registerContainerEvents() (<-chan events.Message, <-chan error) {
	eventFilter := filters.NewArgs()
	eventFilter.Add("type", "container")

	options := types.EventsOptions{
		Filters: eventFilter,
	}
	evtCh, errCh := u.dockerClient.Events(u.ctx, options)

	return evtCh, errCh
}

func (u DNSUpdater) addContainerToDNS(e events.Message) {
	ip, err := u.getContainerIP(e.Actor.ID)
	if err != nil {
		logrus.Errorf("could not determine container ip: %v", err)

		return
	}

	containerName, err := u.getContainerName(e.Actor.ID)
	if err != nil {
		logrus.Errorf("could not determine container name: %v", err)

		return
	}

	logrus.Infof("adding container %s due to (%s) event", containerName, e.Action)

	u.dnsRegistry.Register(containerName, ip)
}

func (u DNSUpdater) removeContainerFromDNS(e events.Message) {
	containerName, err := u.getContainerName(e.Actor.ID)
	if err != nil {
		logrus.Errorf("could not determine container name: %v", err)
		return
	}

	logrus.Infof("removing container %s due to (%s) event", containerName, e.Action)

	u.dnsRegistry.Unregister(containerName)
}

func (u DNSUpdater) getContainerName(containerID string) (string, error) {
	container, err := u.getContainerByID(containerID)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGettingContainerName, err)
	}

	return container.Names[0], nil
}

func (u DNSUpdater) getContainerIP(containerID string) (string, error) {
	container, err := u.getContainerByID(containerID)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGettingContainerIP, err)
	}

	ips := u.dockerClientAdapter.GetContainerNetworkIps(container)

	return ips[len(ips)-1], nil
}

func (u DNSUpdater) getContainerByID(containerID string) (types.Container, error) {
	ctx := context.Background()
	containerFilter := filters.NewArgs()
	containerFilter.Add("id", containerID)

	options := types.ContainerListOptions{
		Filters: containerFilter,
		All:     true,
	}

	containers, err := u.dockerClient.ContainerList(ctx, options)
	if err != nil {
		return types.Container{}, fmt.Errorf("%w using container list: %w", ErrGettingContainerByID, err)
	}

	if len(containers) == 0 {
		return types.Container{},
			fmt.Errorf(
				"%w for id '%s'",
				ErrNoContainerFoundForID,
				containerID,
			)
	}

	return containers[0], nil
}
