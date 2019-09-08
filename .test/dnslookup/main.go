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
	f, err := os.OpenFile(dnsTesterLog, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0655)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writeTestOutput(f, "starting functional test")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), getTimeoutFromEnv())

	testSuccess := make(chan bool)
	go lookup(f, ctx, testSuccess, "pong", "regular docker name, basic docker name resolving, should resolve without of docker-dns")
	go lookup(f, ctx, testSuccess, "www.pong.com", "some custom domain, needs docker-dns")
	go lookup(f, ctx, testSuccess, "ponge.longe.long.com", "some other custom domain, needs docker-dns")
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
			writeTestOutput(f, "%v of %v tests successful", noTestsSuccessful, numberOfTests)
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
		_, err = fmt.Fprintf(f, format+"\n", args...)
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

func lookup(f *os.File, ctx context.Context, testSuccess chan bool, host string, testDescription string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			timeoutCtx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(4))
			_, err := net.DefaultResolver.LookupIPAddr(timeoutCtx, host)
			if err == nil {
				writeTestOutput(f, "success for test input '%s' (%s)", host, testDescription)
				testSuccess <- true
				return
			}
			writeTestOutput(f, "error for test input '%s' (%s): %v", host, testDescription, err)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
