package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"
)

var ctx = context.Background()
var dockerClient *client.Client

const image = "golang:1.12.4"
const containerStopTimeout = time.Second * 10
const containerDir = "/app"
const dockerSocketVolumeBind = "/var/run/docker.sock:/var/run/docker.sock"
const dnsTesterSuccessPattern = "all tests successful"
const dnsTesterOutputFile = "dns-tester.log"
const networkName = "test_network"
const subNet = "172.28.0.0/16"
const ipRange = "172.28.5.0/24"
const dnsServerIP = "172.28.5.1"
const dnsServerAliasEnv = "DOCKER_DNS_ALIAS_FILE=dnsserver/data/alias"
const dnsServerContainerName = "dns-server"
const dnsTesterContainerName = "test"
const pongContainerName = "pong"
const dnsServerCmd = "go run dnsserver/cmd/main.go"
const dnsTesterCmd = "go run .test/dnslookup/main.go"
const pongCmd = "go run .test/pong/main.go"

type Net struct {
	NetworkID   string
	NetworkName string
}

func main() {
	fmt.Println("connecting to docker")
	var err error
	dockerClient, err = client.NewEnvClient()
	panicOnError(err)

	fmt.Println("create network")
	networkInfo := createNetwork()

	fmt.Println("create containers")
	dnsContainerID := createDnsContainer(networkInfo)
	dnsTesterContainerID := createTesterContainer(networkInfo)
	pongContainerID := createSimpleGoContainer(pongContainerName, pongCmd, networkInfo)

	fmt.Println("start containers")
	startContainer(dnsContainerID)
	startContainer(dnsTesterContainerID)
	startContainer(pongContainerID)

	fmt.Println("wait for tests to finish")
	waitForTestFinish(dnsTesterContainerID)

	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	<-sigChannel

	fmt.Println("cleanup")
	cleanup()

	fmt.Println("check test results")
	res := checkResults()
	os.Exit(res)
}

func waitForTestFinish(containerID string) {
	go func() {
		waitContainerToFadeAway(containerID)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()
}

func checkResults() int {
	content, err := ioutil.ReadFile(dnsTesterOutputFile)
	panicOnError(err)

	if strings.Contains(string(content), dnsTesterSuccessPattern) {
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

func waitContainerToFadeAway(containerID string) {
	for {
		_, err := dockerClient.ContainerInspect(ctx, containerID)
		if client.IsErrNotFound(err) {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func cleanup() {
	shutDownContainers := &sync.WaitGroup{}
	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{Filters: getFilterArgs()})
	if err == nil {
		shutDownContainers.Add(len(containers))
		for _, testContainer := range containers {
			go shutDownContainer(testContainer.ID, shutDownContainers)
		}
	} else {
		fmt.Printf("error finding test containers: %v\n", err)
	}
	shutDownContainers.Wait()
	cleanupTestNetwork()
}

func startContainer(containerID string) {
	err := dockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	panicOnError(err)
}

func shutDownContainer(containerID string, wg *sync.WaitGroup) {
	stopTimeout := containerStopTimeout
	_ = dockerClient.ContainerStop(ctx, containerID, &stopTimeout)

	waitContainerToFadeAway(containerID)
	wg.Done()
}

func createSimpleGoContainer(containerName, cmd string, networkInfo Net) string {

	containerConfig, hostConfig, networkConfig := createBaseGoContainerStructs(cmd, networkInfo)

	containerBody, err := dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, containerName)
	panicOnError(err)

	return containerBody.ID
}

func createTesterContainer(networkInfo Net) string {

	containerConfig, hostConfig, networkConfig := createBaseGoContainerStructs(dnsTesterCmd, networkInfo)

	hostConfig.DNS = []string{dnsServerIP}

	containerBody, err := dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, dnsTesterContainerName)
	panicOnError(err)
	testerContainerID := containerBody.ID

	return testerContainerID
}

func createDnsContainer(networkInfo Net) string {

	containerConfig, hostConfig, networkConfig := createBaseGoContainerStructs(dnsServerCmd, networkInfo)

	networkConfig.EndpointsConfig[networkInfo.NetworkName].IPAMConfig = &network.EndpointIPAMConfig{IPv4Address: dnsServerIP}
	hostConfig.Binds = append(hostConfig.Binds, dockerSocketVolumeBind)
	containerConfig.Env = append(containerConfig.Env, dnsServerAliasEnv)

	containerBody, err := dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, dnsServerContainerName)
	panicOnError(err)
	dnsContainerID := containerBody.ID

	return dnsContainerID
}

func createBaseGoContainerStructs(cmd string, networkInfo Net) (*container.Config, *container.HostConfig, *network.NetworkingConfig) {
	hostDir, _ := os.Getwd()
	fmt.Println("binding host volume: ", hostDir, containerDir)

	containerConfig := &container.Config{
		Env:        []string{"GOPROXY=https://proxy.golang.org", "GO111MODULE=on"},
		WorkingDir: containerDir,
		Image:      image,
		Cmd:        strslice.StrSlice(strings.Split(cmd, " ")),
		Labels:     map[string]string{"docker-dns": "functional-test"},
	}

	hostConfig := &container.HostConfig{
		AutoRemove:  true,
		NetworkMode: container.NetworkMode(networkInfo.NetworkName),
		Binds:       []string{hostDir + ":" + containerDir},
	}

	if goPath, ok := os.LookupEnv("GOPATH"); ok {
		hostConfig.Binds = append(hostConfig.Binds, path.Join(goPath, "pkg")+":/go/pkg")
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkInfo.NetworkName: {
				NetworkID: networkInfo.NetworkID,
			},
		},
	}

	return containerConfig, hostConfig, networkConfig
}

func createNetwork() Net {
	cleanupTestNetwork()

	options := types.NetworkCreate{
		CheckDuplicate: true,
		Attachable:     true,
		Driver:         "bridge",
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				{
					Subnet:  subNet,
					IPRange: ipRange,
				},
			},
		},
		Labels: map[string]string{"docker-dns": "functional-test"},
	}

	resp, err := dockerClient.NetworkCreate(ctx, networkName, options)
	panicOnError(err)

	return Net{resp.ID, networkName}
}

func cleanupTestNetwork() {
	res, err := dockerClient.NetworkList(ctx, types.NetworkListOptions{Filters: getFilterArgs()})
	panicOnError(err)
	for _, networkResource := range res {
		err := dockerClient.NetworkRemove(ctx, networkResource.ID)
		if err != nil {
			fmt.Printf("could not remove network: %v\n", err)
		}
	}
}

func getFilterArgs() filters.Args {
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "docker-dns=functional-test")
	return filterArgs
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
