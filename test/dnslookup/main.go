package main

import (
	"context"
	"fmt"
	"io"
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
const filePerm = 0655

func main() {
	os.Exit(run())
}

func run() int {
	f, err := os.OpenFile(dnsTesterLog, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, filePerm)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	writeTestOutputf(f, "starting functional test")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), getTimeoutFromEnv())
	defer cancel()

	testSuccess := make(chan bool)
	go lookup(ctx, f, testSuccess, "pong", "regular docker name, basic docker name resolving, should resolve without of docker-dns") //nolint:lll
	go lookup(ctx, f, testSuccess, "www.pong.com", "some custom domain, needs docker-dns")
	go lookup(ctx, f, testSuccess, "ponge.longe.long.com", "some other custom domain, needs docker-dns")

	var (
		numberOfTests     = 3
		noTestsSuccessful = 0
	)

	for {
		select {
		case sig := <-signals:
			writeTestOutputf(f, "received signal %v", sig)
			cancel()

			return 0
		case <-ctx.Done():
			writeTestOutputf(f, "test timed out")

			return 1
		case <-testSuccess:
			noTestsSuccessful++
			writeTestOutputf(f, "%v of %v tests successful", noTestsSuccessful, numberOfTests)
		default:
			if noTestsSuccessful == numberOfTests {
				writeTestOutputf(f, successMessage)

				return 0
			}
		}
	}
}

func writeTestOutputf(f io.Writer, format string, args ...interface{}) {
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
	}

	return time.Duration(timeout) * time.Second
}

func lookup(ctx context.Context, f io.Writer, testSuccess chan bool, host string, testDescription string) {
	const (
		ipLookupTimeout = time.Second * 4
		delay           = 100 * time.Millisecond
	)

	for {
		select {
		case <-ctx.Done():
			return

		default:
			timeoutCtx, cancel := context.WithTimeout(context.Background(), ipLookupTimeout)

			_, err := net.DefaultResolver.LookupIPAddr(timeoutCtx, host)
			if err == nil {
				writeTestOutputf(f, "success for test input '%s' (%s)", host, testDescription)
				testSuccess <- true

				cancel()

				return
			}

			cancel()

			writeTestOutputf(f, "error for test input '%s' (%s): %v", host, testDescription, err)

			time.Sleep(delay)
		}
	}
}
