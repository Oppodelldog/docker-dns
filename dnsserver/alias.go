package dnsserver

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/Oppodelldog/filediscovery"

	"github.com/sirupsen/logrus"
)

const aliasLoaderDefaultInterval = 10 * time.Second
const aliasFilePathEnvKey = "DOCKER_DNS_ALIAS_FILE"

// AliasFileLoader loads the alias file which holds value pairs defining alias for a container name.
// see data/alias for an example.
type AliasFileLoader struct {
	aliases         map[string]string
	aliasFileFinder filediscovery.FileDiscoverer
	lock            sync.Mutex
}

// NewAliasFileLoader creates a new *AliasFileLoader.
func NewAliasFileLoader(ctx context.Context) *AliasFileLoader {
	a := &AliasFileLoader{
		aliases: map[string]string{},
		aliasFileFinder: filediscovery.New(
			[]filediscovery.FileLocationProvider{
				filediscovery.EnvVarFilePathProvider(aliasFilePathEnvKey),
				filediscovery.ExecutableDirProvider("data"),
			},
		),
		lock: sync.Mutex{},
	}

	a.startAliasLoader(ctx)

	return a
}

func (l *AliasFileLoader) startAliasLoader(ctx context.Context) {
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

func (l *AliasFileLoader) loadAliasesFromFile() {
	logrus.Debug("Loading Alias file")

	aliasFilePath, err := l.aliasFileFinder.Discover("alias")
	if err != nil {
		logrus.Errorf("Could not find alias file: %v\n", err)
		return
	}

	logrus.Infof("Loading Alias file from '%s'", aliasFilePath)

	content, err := ioutil.ReadFile(aliasFilePath)
	if err != nil {
		logrus.Errorf("Could not load aliases: %v\n", err)
		return
	}

	newAliases := map[string]string{}

	scanner := bufio.NewScanner(bytes.NewBuffer(content))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		const requiredValues = 2
		if len(fields) == requiredValues {
			if strings.Contains(fields[0], "#") {
				continue
			}

			newAliases[fields[0]] = fields[1]
		}
	}

	numberOfAliases := len(newAliases)
	logrus.Errorf("number of aliases: %v\n", numberOfAliases)

	if numberOfAliases > 0 {
		l.lock.Lock()
		l.aliases = newAliases
		l.lock.Unlock()
	}
}

func (l *AliasFileLoader) GetAliasForDomain(domain string) (string, bool) {
	if alias, ok := l.aliases[domain]; ok {
		return alias, true
	}

	return "", false
}
