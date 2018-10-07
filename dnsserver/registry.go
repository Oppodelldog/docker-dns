package dnsserver

import (
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

	if val, ok := r.ipAddressByContainerName[domain]; ok {
		return val, ok
	}

	return "", false
}

func (r *dnsRegistry) Unregister(containerName string) {
	containerName = r.normalizeContainerName(containerName)
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.ipAddressByContainerName, containerName)
}

func (r *dnsRegistry) Register(containerName string, ip string) {
	containerName = r.normalizeContainerName(containerName)
	r.lock.Lock()
	defer r.lock.Unlock()

	r.ipAddressByContainerName[containerName] = ip
}

func (r *dnsRegistry) normalizeContainerName(containerName string) string {
	parts := strings.Split(containerName, "_")
	if len(parts) < 2 {
		return containerName
	}

	return parts[1]
}
