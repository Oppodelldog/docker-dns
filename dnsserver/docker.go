package dnsserver

import (
	"context"

	"github.com/Oppodelldog/docker-dns/helper"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type (
	RunningContainersGetter interface {
		GetRunningContainers() ([]types.Container, error)
	}
	NetworkIDsGetter interface {
		GetNetworkIDs() ([]string, error)
	}
	NetworkIPsGetter interface {
		GetContainerNetworkIps(container types.Container) []string
	}
	dockerClientAdapter struct {
		dockerClient *client.Client
	}
)

func NewDockerClientAdapter(dockerClient *client.Client) *dockerClientAdapter {
	return &dockerClientAdapter{
		dockerClient: dockerClient,
	}
}

func (a *dockerClientAdapter) GetRunningContainers() ([]types.Container, error) {
	containers, err := a.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (a *dockerClientAdapter) getNetworkIDs() ([]string, error) {

	var networkIDs []string
	myIps, err := helper.GetIps()
	if err != nil {
		return nil, err
	}
	containers, err := a.GetRunningContainers()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		for _, network := range container.NetworkSettings.Networks {
			for _, ip := range myIps {
				if network.IPAddress == ip.String() {
					networkIDs = append(networkIDs, network.NetworkID)
				}
			}
		}
	}

	return networkIDs, nil
}

func (a *dockerClientAdapter) GetContainerNetworkIps(container types.Container) []string {
	var ips []string
	networkIDs, err := a.getNetworkIDs()
	if err != nil {
		logrus.Errorf("error retrieving all NetworkIDs: %v", err)

		return nil
	}
	for _, network := range container.NetworkSettings.Networks {
		for _, myNetwork := range networkIDs {
			if network.NetworkID == myNetwork {
				ips = append(ips, network.IPAddress)
			}
		}
	}

	return ips
}
