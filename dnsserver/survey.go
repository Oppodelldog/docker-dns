package dnsserver

type ContainerDNSSurvey struct {
	dnsRegisterer          DNSRegisterer
	runningContainerGetter RunningContainersGetter
	networkIPsGetter       NetworkIPsGetter
}

func NewContainerDNSSurvey(dnsRegisterer DNSRegisterer,
	runningContainerGetter RunningContainersGetter,
	networkIPsGetter NetworkIPsGetter) *ContainerDNSSurvey {
	return &ContainerDNSSurvey{
		networkIPsGetter:       networkIPsGetter,
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
		ips := s.networkIPsGetter.GetContainerNetworkIps(container)
		for _, containerName := range container.Names {
			ip := ips[len(ips)-1]
			s.dnsRegisterer.Register(containerName, ip)
		}
	}
}
