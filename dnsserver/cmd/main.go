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

	ctx := getContextCanceledByInterrupt()

	dnsserver.StartDockerDNSSurvey(ctx)
	dnsserver.StartAliasLoader(ctx)
	dnsserver.Run(ctx)
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
