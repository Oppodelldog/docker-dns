package dnsserver

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/miekg/dns"
)

const dnsPort = 53

// Run starts the DNS server which will answer requests using the given IPResolver
func Run(ctx context.Context, ipResolver IPResolver) {

	s := spawnServer(ipResolver)

	<-ctx.Done()

	stopServer(s)
}

func stopServer(s *dns.Server) {
	err := s.Shutdown()
	if err != nil {
		fmt.Printf("Failed to gracefully shutdown udp listener %s\n", err.Error())
		os.Exit(1)
	}
}

func spawnServer(ipResolver IPResolver) *dns.Server {
	fmt.Printf("starting dns server (udp) on :%v\n", dnsPort)
	srv := &dns.Server{Addr: ":" + strconv.Itoa(dnsPort), Net: "udp"}
	srv.Handler = NewDNSHandler(ipResolver)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("Failed to set udp listener %s\n", err.Error())
			os.Exit(1)
		}
	}()

	return srv
}
