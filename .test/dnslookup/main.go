package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const envTimeoutKey = "TIMEOUT_IN_SECONDS"
const defaultTimeout = time.Minute * 2
const dnsTesterLog = "dns-tester.log"
const successMessage = "all tests successful"

func main() {
	f, err := os.Create(dnsTesterLog)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writeTestOutput(f, "starting functional test")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), getTimeoutFromEnv())

	testSuccess := make(chan bool)
	go lookup(ctx, testSuccess, "pong")
	go lookup(ctx, testSuccess, "www.pong.com")
	go lookup(ctx, testSuccess, "ponge.longe.long.com")
	numberOfTests := 3

	var noTestsSuccessful int
	for {
		select {
		case sig := <-signals:
			writeTestOutput(f, "received signal %v", sig)
			cancel()
			os.Exit(0)
		case <-ctx.Done():
			writeTestOutput(f, "test timed out")
			os.Exit(1)
		case <-testSuccess:
			noTestsSuccessful++
			writeTestOutput(f, "partial test successful")
		default:
			if noTestsSuccessful == numberOfTests {
				writeTestOutput(f, successMessage)
				return
			}
		}
	}
}

func writeTestOutput(f *os.File, format string, args ...interface{}) {
	var err error
	if len(args) > 0 {
		_, err = fmt.Fprintf(f, format+"\n", args)
	} else {
		_, err = fmt.Fprint(f, format+"\n")
	}

	if err != nil {
		panic(fmt.Sprintf("could not write to test output: %v", err))
	}
}

func getTimeoutFromEnv() time.Duration {
	sVal := os.Getenv(envTimeoutKey)
	timeout, err := strconv.Atoi(sVal)
	if err != nil {
		return defaultTimeout
	} else {
		return time.Duration(timeout) * time.Second
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
			fmt.Println(err)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
