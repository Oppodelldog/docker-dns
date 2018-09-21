package dnsserver

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/Oppodelldog/docker-dns/helper"
)

const aliasLoaderDefaultInterval = 10 * time.Second
const aliasFileDefaultPath = "data/alias"

var aliases = map[string]string{}
var ipAddressByContainerName = map[string]string{}
var lock sync.Mutex

//StartAliasLoader starts frequently loading of the alias file
func StartAliasLoader(ctx context.Context) {
	go func() {
		loadAliases()
		ticker := time.NewTicker(aliasLoaderDefaultInterval)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Stopping Alias loader")
				return
			case <-ticker.C:
				loadAliases()
			}
		}
	}()
}

func getIPFromDomain(domain string) (string, bool) {
	lock.Lock()
	defer lock.Unlock()

	if alias, ok := aliases[domain]; ok {
		domain = alias
	}

	if val, ok := ipAddressByContainerName[domain]; ok {
		return val, ok
	}

	return "", false
}

func updateIPContainerNameMappings(m map[string]string) {
	lock.Lock()
	defer lock.Unlock()

	ipAddressByContainerName = m
}

func loadAliases() {
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
			newAliases[makeDNSDomain(fields)] = fields[1]
		}
	}

	numberOfAliases := len(newAliases)

	if numberOfAliases > 0 {
		lock.Lock()
		defer lock.Unlock()

		aliases = newAliases
		helper.PrintStringMap(aliases)
	}
}

func makeDNSDomain(fields []string) string {
	return fmt.Sprintf("%s.", fields[0])
}
