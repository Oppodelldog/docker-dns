package dnsserver

import (
	"fmt"
	"strings"
)

type ContainerDNSSurvey struct {
	dnsRegisterer          DNSRegisterer
	runningContainerGetter RunningContainersGetter
}

func NewContainerDNSSurvey(dnsRegisterer DNSRegisterer,
	runningContainerGetter RunningContainersGetter) *ContainerDNSSurvey {
	return &ContainerDNSSurvey{
		dnsRegisterer:          dnsRegisterer,
		runningContainerGetter: runningContainerGetter,
	}
}

func (s *ContainerDNSSurvey) Run() {
	containers, err := s.runningContainerGetter.GetRunningContainers()
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		ips := GetContainerNetworkIps(container)
		for _, containerName := range container.Names {
			ip := ips[len(ips)-1]
			//TODO: why is docker-compose name necessary here? this service should work for docker also
			dockerComposeContainerName := s.getDockerComposeName(containerName)
			s.dnsRegisterer.Register(dockerComposeContainerName, ip)
		}
	}
}

func (s *ContainerDNSSurvey) getDockerComposeName(containerName string) string {
	parts := strings.Split(containerName, "_")
	if len(parts) < 2 {
		fmt.Println("could not convert to docker-compose name")
		return containerName
	}

	return parts[1]
}
