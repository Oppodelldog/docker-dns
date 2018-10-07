package dnsserver

import (
	"context"
	"fmt"
	"github.com/Oppodelldog/docker-dns/helper"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type RunningContainersGetter interface {
	GetRunningContainers() ([]types.Container, error)
}

type RunningContainersGetterFunc func() ([]types.Container, error)

func (f RunningContainersGetterFunc) GetRunningContainers() ([]types.Container, error) {
	return f()
}

type NetworkIDsGetter interface {
	GetNetworkIDs() ([]string, error)
}

func GetRunningContainers() ([]types.Container, error) {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = dockerClient.Close()
		if err != nil {
			fmt.Printf("error cloding docker dockerClient: %v", err)
		}
	}()

	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func getNetworkIDs() ([]string, error) {

	var networkIDs []string
	myIps, err := helper.GetIps()
	if err != nil {
		return nil, err
	}
	containers, err := GetRunningContainers()
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

func GetContainerNetworkIps(container types.Container) []string {
	var ips []string
	networkIDs, err := getNetworkIDs()
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
