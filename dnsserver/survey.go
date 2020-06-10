package dnsserver

import (
	"github.com/sirupsen/logrus"
)

type ContainerDNSSurvey struct {
	dnsRegisterer          DNSRegisterer
	runningContainerGetter RunningContainersGetter
	networkIPsGetter       NetworkIPsGetter
}

func NewContainerDNSSurvey(dnsRegisterer DNSRegisterer,
	runningContainerGetter RunningContainersGetter,
	networkIPsGetter NetworkIPsGetter) ContainerDNSSurvey {
	return ContainerDNSSurvey{
		networkIPsGetter:       networkIPsGetter,
		dnsRegisterer:          dnsRegisterer,
		runningContainerGetter: runningContainerGetter,
	}
}

func (s ContainerDNSSurvey) Run() {
	containers, err := s.runningContainerGetter.GetRunningContainers()
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		ips := s.networkIPsGetter.GetContainerNetworkIps(container)
		if len(ips) == 0 {
			logrus.Debugf("skipping container without ip '%s'", container.ID)
			continue
		}

		for _, containerName := range container.Names {
			ip := ips[len(ips)-1]
			s.dnsRegisterer.Register(containerName, ip)
		}
	}
}
