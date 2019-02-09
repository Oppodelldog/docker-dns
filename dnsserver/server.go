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

type dnsHandler struct {
	ipResolver IPResolver
}

func newDNSHandler(ipResolver IPResolver) dns.Handler {
	return &dnsHandler{
		ipResolver: ipResolver,
	}
}

// Run starts the DNS server which will answer requests using the given IPResolver
func Run(ctx context.Context, ipResolver IPResolver) {

	s := spawnServer(ipResolver)

	<-ctx.Done()

	stopServer(s)
}

func stopServer(s *dns.Server) {
	err := s.Shutdown()
	if err != nil {
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

//ServeDNS handles a dns request
func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name

		address, ok := h.ipResolver.LookupIP(domain)
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		}
	}

	err := w.WriteMsg(&msg)
	if err != nil {
		logrus.Errorf("Error writing DNS response: %v", err)
	}
}
