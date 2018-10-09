package dnsserver

import (
	"fmt"
	"strings"
	"sync"
)

type (
	DNSRegisterer interface {
		Register(domain string, ip string)
	}
	DNSUnRegisterer interface {
		Unregister(domain string)
	}
	DNSRegistrar interface {
		DNSRegisterer
		DNSUnRegisterer
	}
	IPResolver interface {
		LookupIP(string) (string, bool)
	}
	DNSRegistry interface {
		DNSRegisterer
		DNSUnRegisterer
		IPResolver
	}
)

//NewDNSRegistry returns a new instance of DNRegistry
func NewDNSRegistry(aliasProvider AliasProvider) DNSRegistry {
	return &dnsRegistry{
		ipAddressByContainerName: map[string]string{},
		lock:          sync.Mutex{},
		aliasProvider: aliasProvider,
	}
}

type (
	dnsRegistry struct {
		ipAddressByContainerName map[string]string
		lock                     sync.Mutex
		aliasProvider            AliasProvider
	}
	AliasProvider interface {
		GetAliasForDomain(string) (string, bool)
	}
)

func (r *dnsRegistry) LookupIP(domain string) (string, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if alias, ok := r.aliasProvider.GetAliasForDomain(domain); ok {
		domain = alias
	}

	if ip, ok := r.ipAddressByContainerName[domain]; ok {
		return ip, ok
	}

	return "", false
}

func (r *dnsRegistry) Unregister(containerName string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.ipAddressByContainerName, containerName)
}

func (r *dnsRegistry) Register(containerName string, ip string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.ipAddressByContainerName[containerName] = ip
}

func NewContainerRegistry(registerer DNSRegistrar) DNSRegistrar {
	return &ContainerDNSRegistry{
		registry: registerer,
	}
}

type ContainerDNSRegistry struct {
	registry DNSRegistrar
}

func (r *ContainerDNSRegistry) Unregister(containerName string) {
	dnsContainerName := r.normalizeContainerName(containerName)

	r.registry.Unregister(dnsContainerName)
}

func (r *ContainerDNSRegistry) Register(containerName string, ip string) {
	dnsContainerName := r.normalizeContainerName(containerName)

	r.registry.Register(dnsContainerName, ip)
}
func (r *ContainerDNSRegistry) normalizeContainerName(containerName string) string {
	var dnsContainerName string
	parts := strings.Split(containerName, "_")
	if len(parts) < 2 {
		if containerName[0] == '/' {
			containerName = containerName[1:]
		}
		dnsContainerName = containerName
	} else {
		dnsContainerName = parts[1]
	}

	return fmt.Sprintf("%s.", dnsContainerName)
}
