package helper

import (
	"fmt"
	"net"
)

// PrintIps just prints the programs systems ip addresses to stdout
func PrintIps() {
	ips, err := GetIps()
	if err != nil {
		fmt.Printf("unable to determine ip addresses: %v", err)
		return
	}

	for _, ip := range ips {
		fmt.Println(ip)
	}
}

// GetIps gets systems ip addresses
func GetIps() ([]net.IP, error) {
	var ips []net.IP
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {

			switch v := addr.(type) {
			case *net.IPNet:
				ips = append(ips, v.IP)
			case *net.IPAddr:
				ips = append(ips, v.IP)
			}
		}
	}

	return ips, nil
}

//PrintStringMap awesomely prints a string map with an awesome arrow
func PrintStringMap(m map[string]string) {
	for k, v := range m {
		fmt.Printf("%s -> %s\n", k, v)
	}
}
