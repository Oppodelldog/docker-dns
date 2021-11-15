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
)

// NewDNSRegistry returns a new instance of DNSRegistry.
func NewDNSRegistry(aliasProvider AliasProvider) DNSRegistry {
	return DNSRegistry{
		ipAddressByContainerName: map[string]string{},
		lock:                     &sync.Mutex{},
		aliasProvider:            aliasProvider,
	}
}

type (
	DNSRegistry struct {
		ipAddressByContainerName map[string]string
		lock                     *sync.Mutex
		aliasProvider            AliasProvider
	}
	AliasProvider interface {
		GetAliasForDomain(string) (string, bool)
	}
)

func (r DNSRegistry) LookupIP(domain string) (string, bool) {
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

func (r DNSRegistry) Unregister(containerName string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.ipAddressByContainerName, containerName)
}

func (r DNSRegistry) Register(containerName string, ip string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.ipAddressByContainerName[containerName] = ip
}

// NewContainerRegistry creates a new instance of ContainerDNSRegistry.
func NewContainerRegistry(registerer DNSRegistrar) ContainerDNSRegistry {
	return ContainerDNSRegistry{
		registry: registerer,
	}
}

type ContainerDNSRegistry struct {
	registry DNSRegistrar
}

func (r ContainerDNSRegistry) Unregister(containerName string) {
	dnsContainerName := r.normalizeContainerName(containerName)

	r.registry.Unregister(dnsContainerName)
}

func (r ContainerDNSRegistry) Register(containerName string, ip string) {
	dnsContainerName := r.normalizeContainerName(containerName)

	r.registry.Register(dnsContainerName, ip)
}

func (r ContainerDNSRegistry) normalizeContainerName(containerName string) string {
	var dnsContainerName string

	parts := strings.Split(containerName, "_")

	const requiredValues = 2
	if len(parts) < requiredValues {
		if containerName[0] == '/' {
			containerName = containerName[1:]
		}

		dnsContainerName = containerName
	} else {
		dnsContainerName = parts[1]
	}

	return fmt.Sprintf("%s.", dnsContainerName)
}
