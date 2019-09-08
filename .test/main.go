package main

import (
	"context"
	"fmt"
	"github.com/Oppodelldog/dockertest"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

const goImage = "golang:1.13.0"
const imagePackagePath = "/go/pkg"
const containerProjectRoot = "/go/src/github.com/Oppodelldog"
const workingDir = "/go/src/github.com/Oppodelldog/docker-dns"

const subNet = "172.28.0.0/16"
const ipRange = "172.28.5.0/24"
const dnsServerIP = "172.28.5.1"
const networkName = "test_network"

const dockerSocketPath = "/var/run/docker.sock"

const dnsTesterOutputFile = "dns-tester.log"

var ctx = context.Background()

func main() {
	hostDir, _ := os.Getwd()
	projectRoot := filepath.Dir(hostDir)
	goPath, hasGoPath := os.LookupEnv("GOPATH")
	if !hasGoPath {
		panic("did not find GOPATH, it's required for caching modules")
	}
	localPackagePath := path.Join(goPath, "pkg")

	fmt.Println("connecting to docker")
	var err error
	dt, err := dockertest.New()
	if err != nil {
		panic(err)
	}

	fmt.Println("create network")
	networkBuilder := dt.CreateSimpleNetwork(networkName, subNet, ipRange)
	net, err := networkBuilder.Create()
	panicOnError(err)

	fmt.Println("create containers")

	dnsContainer, err := dt.NewContainer("dns-server", goImage, "go run dnsserver/cmd/main.go").
		ConnectToNetwork(net).
		SetIPAddress(dnsServerIP, networkName).
		Mount(localPackagePath, imagePackagePath).
		Mount(projectRoot, containerProjectRoot).
		Mount(dockerSocketPath, dockerSocketPath).
		SetEnv("DOCKER_DNS_ALIAS_FILE", "dnsserver/data/alias").
		SetWorkingDir(workingDir).
		CreateContainer()
	panicOnError(err)

	dnsTesterContainer, err := dt.NewContainer("test", goImage, "go run .test/dnslookup/main.go").
		ConnectToNetwork(net).
		AddDns(dnsServerIP).
		Mount(localPackagePath, imagePackagePath).
		Mount(projectRoot, containerProjectRoot).
		SetWorkingDir(workingDir).
		CreateContainer()
	panicOnError(err)

	pongContainer, err := dt.NewContainer("pong", goImage, "go run .test/pong/main.go").
		ConnectToNetwork(net).
		Mount(localPackagePath, imagePackagePath).
		Mount(projectRoot, containerProjectRoot).
		SetWorkingDir(workingDir).
		CreateContainer()
	panicOnError(err)

	fmt.Println("start containers")
	err = dnsContainer.StartContainer()
	panicOnError(err)
	err = pongContainer.StartContainer()
	panicOnError(err)
	err = dnsTesterContainer.StartContainer()
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
