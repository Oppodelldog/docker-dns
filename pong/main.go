package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Oppodelldog/docker-dns/helper"
)

func main() {
	fmt.Println("I am pong, a long living process that does nothing, just keep the container running which is used as dns lookup target")
	helper.PrintIps()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	sig := <-signals
	fmt.Println("signal", sig)

	os.Exit(0)

}
