package main

import (
	"context"
	"fmt"
	"github.com/Oppodelldog/dockertest"
	"github.com/docker/docker/api/types/network"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const image = "golang:1.13.0"
const containerDir = "/app"
const subNet = "172.28.0.0/16"
const ipRange = "172.28.5.0/24"
const dnsServerIP = "172.28.5.1"
const dnsTesterOutputFile = "dns-tester.log"
const networkName = "test_network"

var ctx = context.Background()

func main() {
	hostDir, _ := os.Getwd()

	fmt.Println("connecting to docker")
	var err error
	dt, err := dockertest.New()
	if err != nil {
		panic(err)
	}

	fmt.Println("create network")
	dt.CreateNetwork(networkName, subNet, ipRange)

	fmt.Println("create containers")
	dnsContainerBuilder := dt.NewContainer("dns-server", image, "go run dnsserver/cmd/main.go")
	dnsContainerBuilder.NetworkingConfig.EndpointsConfig[networkName].IPAMConfig = &network.EndpointIPAMConfig{IPv4Address: dnsServerIP}
	dnsContainerBuilder.HostConfig.Binds = []string{hostDir + ":" + containerDir}
	dnsContainerBuilder.HostConfig.Binds = append(dnsContainerBuilder.HostConfig.Binds, "/var/run/docker.sock:/var/run/docker.sock")

	dnsContainerBuilder.ContainerConfig.Env = append(dnsContainerBuilder.ContainerConfig.Env, "DOCKER_DNS_ALIAS_FILE=dnsserver/data/alias")
	dnsContainerBuilder.ContainerConfig.WorkingDir = containerDir

	dnsContainer, err := dnsContainerBuilder.CreateContainer()
	panicOnError(err)

	dnsTesterContainerBuilder := dt.NewContainer("test", image, "go run .test/dnslookup/main.go")
	dnsTesterContainerBuilder.HostConfig.DNS = []string{dnsServerIP}
	dnsTesterContainerBuilder.HostConfig.Binds = []string{hostDir + ":" + containerDir}
	dnsTesterContainerBuilder.ContainerConfig.WorkingDir = containerDir

	dnsTesterContainer, err := dnsTesterContainerBuilder.CreateContainer()
	panicOnError(err)

	pongContainerBuilder := dt.NewContainer("pong", image, "go run .test/pong/main.go")
	pongContainerBuilder.ContainerConfig.WorkingDir = containerDir
	pongContainerBuilder.HostConfig.Binds = []string{hostDir + ":" + containerDir}

	pongContainer, err := pongContainerBuilder.CreateContainer()
	panicOnError(err)

	fmt.Println("start containers")
	err = dnsContainer.StartContainer()
	panicOnError(err)
	err = dnsTesterContainer.StartContainer()
	panicOnError(err)
	err = pongContainer.StartContainer()
	panicOnError(err)

	fmt.Println("wait for tests to finish")
	dt.WaitForContainerToExit(dnsTesterContainer)

	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	<-sigChannel

	dt.DumpContainerLogs(ctx, dnsContainer)
	dt.DumpContainerLogs(ctx, dnsTesterContainer)
	dt.DumpContainerLogs(ctx, pongContainer)

	fmt.Println("cleanup")
	dt.Cleanup()

	fmt.Println("check test results")
	res := checkResults()
	os.Exit(res)
}

func checkResults() int {
	content, err := ioutil.ReadFile(dnsTesterOutputFile)
	if err != nil {
		panic(err)
	}

	if strings.Contains(string(content), "all tests successful") {
		fmt.Println("Test successfull")
		fmt.Println("-------------------------------")
		fmt.Println(string(content))
		return 0
	} else {
		fmt.Println("Test failed")
		fmt.Println("-------------------------------")
		fmt.Println(string(content))
		return 1
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
