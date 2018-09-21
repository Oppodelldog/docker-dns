package dnsserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Oppodelldog/docker-dns/helper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const dokerDNSSurveyDefaultInterval = 10 * time.Second

//StartDockerDNSSurvey frequently scans the running docker containers which are in the same network as this server.
func StartDockerDNSSurvey(ctx context.Context) {
	go func() {
		myNetworks := getNetworkIDs()
		dnsSurvey(myNetworks)
		ticker := time.NewTicker(dokerDNSSurveyDefaultInterval)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Stopping Docker DNS Survey")
				return
			case <-ticker.C:
				dnsSurvey(myNetworks)
			}
		}
	}()
}

func getNetworkIDs() []string {

	var myNetworks []string
	myIps, err := helper.GetIps()
	if err != nil {
		panic(err)
	}
	containers, err := getRunningContainers()
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		for _, network := range container.NetworkSettings.Networks {
			for _, ip := range myIps {
				if network.IPAddress == ip.String() {
					myNetworks = append(myNetworks, network.NetworkID)
				}
			}
		}
	}
	return myNetworks
}

func dnsSurvey(strings []string) {
	containers, err := getRunningContainers()
	if err != nil {
		panic(err)
	}

	ipByContainerName := map[string]string{}
	for _, container := range containers {
		ips := getContainerNetworkIps(container, strings)
		for _, containerName := range container.Names {
			ipByContainerName[getDockerComposeName(containerName)] = ips[len(ips)-1]
		}
	}

	updateIPContainerNameMappings(ipByContainerName)

	fmt.Println("New DNS Survey:")
	helper.PrintStringMap(ipByContainerName)
}

func getDockerComposeName(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) < 2 {
		fmt.Println("could not convert to docker-compose name")
		return s
	}

	return parts[1]
}

func getContainerNetworkIps(container types.Container, myNetworkIds []string) []string {
	var ips []string
	for _, network := range container.NetworkSettings.Networks {
		for _, myNetwork := range myNetworkIds {
			if network.NetworkID == myNetwork {
				ips = append(ips, network.IPAddress)
			}
		}
	}

	return ips
}

func getRunningContainers() ([]types.Container, error) {
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
