package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	"github.com/Oppodelldog/docker-dns/dnsserver"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	ctx := getContextCanceledByInterrupt()

	dockerClient, dockerClientDefer := getDockerClient()
	defer dockerClientDefer()

	dockerClientAdapter := dnsserver.NewDockerClientAdapter(dockerClient)

	aliasProvider := dnsserver.NewAliasFileLoader(ctx)
	dnsRegistry := dnsserver.NewDNSRegistry(aliasProvider)

	containerRegisterer := dnsserver.NewContainerRegistry(dnsRegistry)

	dnsserver.NewContainerDNSSurvey(containerRegisterer, dockerClientAdapter, dockerClientAdapter).Run()
	dnsserver.NewDNSUpdater(ctx, dockerClient, dockerClientAdapter, containerRegisterer)
	dnsserver.Run(ctx, dnsRegistry)
}

func getDockerClient() (*client.Client, func()) {
	dockerClient, err := client.NewEnvClient()

	return dockerClient, func() {
		err = dockerClient.Close()
		if err != nil {
			logrus.Errorf("error closing docker dockerClient: %v", err)
		}
	}
}

func getContextCanceledByInterrupt() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		<-signals
		cancel()
	}()

	return ctx
}
