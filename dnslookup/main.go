package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var wg sync.WaitGroup

func main() {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case sig := <-signals:
			fmt.Println("signal", sig)
			os.Exit(0)
		default:
			go lookup("pong")
			go lookup("www.pong.com")
			go lookup("ponge.longe.long.com")
			wg.Wait()
			time.Sleep(1 * time.Second)
		}
	}
}

func lookup(host string) {
	wg.Add(1)
	ips, err := net.LookupIP(host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
	}
	for _, ip := range ips {
		fmt.Printf("%s. IN A %s\n", host, ip.String())
	}
	wg.Done()
}
