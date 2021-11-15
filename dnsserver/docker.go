package dnsserver

import (
	"context"
	"fmt"

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
	DockerClientAdapter struct {
		dockerClient *client.Client
	}
)

// NewDockerClientAdapter returns a new DockerClientAdapter.
func NewDockerClientAdapter(dockerClient *client.Client) DockerClientAdapter {
	return DockerClientAdapter{
		dockerClient: dockerClient,
	}
}

// GetRunningContainers returns a list of running containers.
func (a DockerClientAdapter) GetRunningContainers() ([]types.Container, error) {
	containers, err := a.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("cannot get running containers: %w", err)
	}

	return containers, nil
}

func (a DockerClientAdapter) getNetworkIDs() ([]string, error) {
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

func (a DockerClientAdapter) GetContainerNetworkIps(container types.Container) []string {
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
