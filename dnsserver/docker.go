package dnsserver

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
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
	myIps, err := getIps()
	if err != nil {
		return nil, err
	}
	containers, err := a.GetRunningContainers()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		for _, containerNetwork := range container.NetworkSettings.Networks {
			for _, ip := range myIps {
				if containerNetwork.IPAddress == ip.String() {
					networkIDs = append(networkIDs, containerNetwork.NetworkID)
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
	for _, containerNetwork := range container.NetworkSettings.Networks {
		for _, myNetwork := range networkIDs {
			if containerNetwork.NetworkID == myNetwork {
				ips = append(ips, containerNetwork.IPAddress)
			}
		}
	}

	return ips
}
