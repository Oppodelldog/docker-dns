package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Oppodelldog/dockertest"
)

const goImage = "golang:1.14.4"
const imagePackagePath = "/go/pkg"
const containerProjectRoot = "/go/src/github.com/Oppodelldog"
const workingDir = "/go/src/github.com/Oppodelldog/docker-dns"

const subNet = "172.28.0.0/16"
const ipRange = "172.28.5.0/24"
const dnsServerIP = "172.28.5.1"
const networkName = "test_network"

const dockerSocketPath = "/var/run/docker.sock"

const dnsTesterOutputFile = "dns-tester.log"

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
	dt, err := dockertest.NewSession()
	if err != nil {
		panic(err)
	}
	dt.SetLogDir(path.Join(hostDir, ".test", "logs"))

	fmt.Println("create network")
	networkBuilder := dt.CreateSimpleNetwork(networkName, subNet, ipRange)
	net, err := networkBuilder.Create()
	panicOnError(err)

	fmt.Println("create containers")

	baseBuilder := dt.NewContainerBuilder().
		Connect(net).
		Mount(localPackagePath, imagePackagePath).
		Mount(projectRoot, containerProjectRoot).
		WorkingDir(workingDir).
		Image(goImage)

	dnsContainer, err := baseBuilder.NewContainerBuilder().
		Name("dns-server").
		Cmd("go run dnsserver/cmd/main.go").
		IPAddress(dnsServerIP, net).
		Mount(dockerSocketPath, dockerSocketPath).
		Env("DOCKER_DNS_ALIAS_FILE", "dnsserver/data/alias").
		Build()
	panicOnError(err)

	dnsTesterContainer, err := baseBuilder.NewContainerBuilder().
		Name("test").
		Cmd("go run .test/dnslookup/main.go").
		Dns(dnsServerIP).
		Build()
	panicOnError(err)

	pongContainer, err := baseBuilder.NewContainerBuilder().
		Name("pong").
		Cmd("go run .test/pong/main.go").
		UseOriginalName().
		Build()
	panicOnError(err)

	fmt.Println("start containers")
	err = dt.StartContainer(dnsContainer, pongContainer, dnsTesterContainer)

	dt.DumpInspect(dnsContainer, pongContainer, dnsTesterContainer)

	fmt.Println("wait for tests to finish")
	<-dt.WaitForContainerToExit(dnsTesterContainer, time.Second*20)

	dt.DumpContainerLogs(dnsContainer, dnsTesterContainer, pongContainer)

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
