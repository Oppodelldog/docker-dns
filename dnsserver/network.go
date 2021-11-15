package dnsserver

import (
	"fmt"
	"net"
)

func getIps() ([]net.IP, error) {
	var ips []net.IP

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("cannot get ips: %w", err)
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, fmt.Errorf("cannot get addrs: %w", err)
		}

		for i := range addrs {
			switch v := addrs[i].(type) {
			case *net.IPNet:
				ips = append(ips, v.IP)
			case *net.IPAddr:
				ips = append(ips, v.IP)
			}
		}
	}

	return ips, nil
}
