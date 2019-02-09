package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)

	testSuccess := make(chan bool)
	go lookup(ctx, testSuccess, "pong")
	go lookup(ctx, testSuccess, "www.pong.com")
	go lookup(ctx, testSuccess, "ponge.longe.long.com")
	numberOfTests := 3

	var noTestsSuccessful int
	for {
		select {
		case sig := <-signals:
			fmt.Println("signal", sig)
			cancel()
			os.Exit(0)
		case <-ctx.Done():
			fmt.Println("test timed out")
			os.Exit(1)
		case <-testSuccess:
			noTestsSuccessful++
		default:
			if noTestsSuccessful == numberOfTests {
				fmt.Println("test successful")
				return
			}
		}
	}
}

func lookup(ctx context.Context, testSuccess chan bool, host string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := net.LookupIP(host)
			if err == nil {
				testSuccess <- true
				return
			}
		}
	}
}
