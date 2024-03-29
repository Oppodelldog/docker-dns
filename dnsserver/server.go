package dnsserver

import (
	"context"
	"net"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/miekg/dns"
)

const dnsPort = 53

type DNSHandler struct {
	ipResolver IPResolver
}

func newDNSHandler(ipResolver IPResolver) DNSHandler {
	return DNSHandler{
		ipResolver: ipResolver,
	}
}

// ServeDNS handles a dns request.
func (h DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	if r.Question[0].Qtype == dns.TypeA {
		msg.Authoritative = true
		domain := msg.Question[0].Name

		address, ok := h.ipResolver.LookupIP(domain)
		if ok {
			logrus.Debugf("address found for %s", domain)

			const ttl = 60

			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl},
				A:   net.ParseIP(address),
			})
		} else {
			logrus.Debugf("address not found for %s", domain)
		}
	}

	if err := w.WriteMsg(&msg); err != nil {
		logrus.Errorf("Error writing DNS response: %v", err)
	}
}

// Run starts the DNS server which will answer requests using the given IPResolver.
func Run(ctx context.Context, ipResolver IPResolver) {
	s := spawnServer(ipResolver)

	<-ctx.Done()

	stopServer(s)
}

func stopServer(s *dns.Server) {
	if err := s.Shutdown(); err != nil {
		logrus.Errorf("Failed to gracefully shutdown udp listener %s\n", err.Error())
		os.Exit(1)
	}
}

func spawnServer(ipResolver IPResolver) *dns.Server {
	logrus.Infof("starting dns server (udp) on :%v\n", dnsPort)

	srv := &dns.Server{Addr: ":" + strconv.Itoa(dnsPort), Net: "udp"}
	srv.Handler = newDNSHandler(ipResolver)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logrus.Errorf("Failed to set udp listener %s\n", err.Error())
			os.Exit(1)
		}
	}()

	return srv
}
