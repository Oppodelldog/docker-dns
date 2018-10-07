package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Oppodelldog/docker-dns/dnsserver"
	"github.com/Oppodelldog/docker-dns/helper"
)

func main() {
	helper.PrintIps()

	dnsRegistry := dnsserver.NewDNSRegistry()

	ctx := getContextCanceledByInterrupt()

	runningContainersGetter := dnsserver.RunningContainersGetterFunc(dnsserver.GetRunningContainers)

	dnsserver.StartAliasLoader(ctx, dnsRegistry)
	dnsserver.NewContainerDNSSurvey(dnsRegistry, runningContainersGetter).Run()
	dnsserver.NewDNSUpdater().Start(ctx, dnsRegistry)
	dnsserver.Run(ctx, dnsRegistry)
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
