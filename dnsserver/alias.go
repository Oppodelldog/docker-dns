package dnsserver

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const aliasLoaderDefaultInterval = 10 * time.Second
const aliasFileDefaultPath = "data/alias"

type aliasFileLoader struct {
	aliases map[string]string
	lock    sync.Mutex
}

func NewAliasFileLoader(ctx context.Context) *aliasFileLoader {
	a := &aliasFileLoader{
		aliases: map[string]string{},
		lock:    sync.Mutex{},
	}

	a.startAliasLoader(ctx)

	return a
}

func (l *aliasFileLoader) startAliasLoader(ctx context.Context) {
	logrus.Info("Starting Alias loader")

	go func() {
		l.loadAliasesFromFile()
		ticker := time.NewTicker(aliasLoaderDefaultInterval)
		for {
			select {
			case <-ctx.Done():
				logrus.Info("Stopping Alias loader")
				return
			case <-ticker.C:
				l.loadAliasesFromFile()
			}
		}
	}()
}

func (l *aliasFileLoader) loadAliasesFromFile() {
	logrus.Info("Loading Alias file")
	content, err := ioutil.ReadFile(aliasFileDefaultPath)
	if err != nil {
		logrus.Errorf("Could not load aliases: %v\n", err)
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
			newAliases[fields[0]] = fields[1]
		}
	}

	numberOfAliases := len(newAliases)

	if numberOfAliases > 0 {
		l.lock.Lock()
		l.aliases = newAliases
		l.lock.Unlock()
	}
}

func (l *aliasFileLoader) GetAliasForDomain(domain string) (string, bool) {
	if alias, ok := l.aliases[domain]; ok {
		return alias, true
	}

	return "", false
}
