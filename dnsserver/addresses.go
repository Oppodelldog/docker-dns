package dnsserver

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/Oppodelldog/docker-dns/helper"
)

const aliasLoaderDefaultInterval = 10 * time.Second
const aliasFileDefaultPath = "data/alias"

//StartAliasLoader starts frequently loading on the given AliasStorage
func StartAliasLoader(ctx context.Context, aliasFileLoader AliasFileLoader) {
	logrus.Info("Starting Alias loader")

	go func() {
		aliasFileLoader.LoadAliasesFromFile()
		ticker := time.NewTicker(aliasLoaderDefaultInterval)
		for {
			select {
			case <-ctx.Done():
				logrus.Info("Stopping Alias loader")
				return
			case <-ticker.C:
				aliasFileLoader.LoadAliasesFromFile()
			}
		}
	}()
}

func NewDNSRegistry() DNSRegistry {
	return &dnsRegistry{
		aliases:                  map[string]string{},
		ipAddressByContainerName: map[string]string{},
		lock: sync.Mutex{},
	}
}

type dnsRegistry struct {
	aliases                  map[string]string
	ipAddressByContainerName map[string]string
	lock                     sync.Mutex
}

func (s *dnsRegistry) LoadAliasesFromFile() {
	fmt.Println("Loading Alias file")
	content, err := ioutil.ReadFile(aliasFileDefaultPath)
	if err != nil {
		fmt.Printf("Could not load aliases: %v\n", err)
		return
	}
	newAliases := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewBuffer(content))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 2 {
			if strings.Contains(fields[0], "#") {
				continue
			}
			newAliases[s.makeDNSDomain(fields)] = fields[1]
		}
	}

	numberOfAliases := len(newAliases)

	if numberOfAliases > 0 {
		s.lock.Lock()
		defer s.lock.Unlock()

		s.aliases = newAliases
		helper.PrintStringMap(s.aliases)
	}
}

func (s *dnsRegistry) makeDNSDomain(fields []string) string {
	return fmt.Sprintf("%s.", fields[0])
}

func (s *dnsRegistry) GetIPFromDomain(domain string) (string, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if alias, ok := s.aliases[domain]; ok {
		domain = alias
	}

	if val, ok := s.ipAddressByContainerName[domain]; ok {
		return val, ok
	}

	return "", false
}

func (s *dnsRegistry) updateIPContainerNameMappings(m map[string]string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ipAddressByContainerName = m
}

func (s *dnsRegistry) Unregister(containerName string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.ipAddressByContainerName, containerName)
}

func (s *dnsRegistry) Register(containerName string, ip string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ipAddressByContainerName[containerName] = ip
}
