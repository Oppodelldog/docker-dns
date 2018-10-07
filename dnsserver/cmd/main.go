package main

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"

	"os"
	"os/signal"
	"syscall"

	"github.com/Oppodelldog/docker-dns/dnsserver"
)

func main() {
	ctx := getContextCanceledByInterrupt()

	dockerClient, dockerClientDefer := getDockerClient()
	defer dockerClientDefer()

	aliasProvider := dnsserver.NewAliasFileLoader(ctx)
	dnsRegistry := dnsserver.NewDNSRegistry(aliasProvider)

	dockerClientAdapter := dnsserver.NewDockerClientAdapter(dockerClient)

	dnsserver.NewContainerDNSSurvey(dnsRegistry, dockerClientAdapter, dockerClientAdapter).Run()
	dnsserver.NewDNSUpdater(ctx, dockerClient, dockerClientAdapter, dnsRegistry)
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
