package finder

import (
	"errors"
	"os"
	"strings"

	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
)

var ErrConfigFileNotFound = errors.New("configuration file isn't found")

func getGlobalConfigFilePaths() []string {
	src := strings.Split(os.Getenv("AQUA_GLOBAL_CONFIG"), ":")
	paths := make([]string, 0, len(src))
	for _, s := range src {
		if s == "" {
			continue
		}
		paths = append(paths, s)
	}
	return paths
}

type configFinder struct{}

type ConfigFinder interface {
	Find(wd, configFilePath string) (string, error)
	Finds(wd, configFilePath string) []string
	GetGlobalConfigFilePaths() []string
}

func NewConfigFinder() ConfigFinder {
	return &configFinder{}
}

func (finder *configFinder) GetGlobalConfigFilePaths() []string {
	return getGlobalConfigFilePaths()
}

func (finder *configFinder) Find(wd, configFilePath string) (string, error) {
	if configFilePath != "" {
		return configFilePath, nil
	}
	configFilePath = findconfig.Find(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
	if configFilePath != "" {
		return configFilePath, nil
	}
	for _, p := range finder.GetGlobalConfigFilePaths() {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		return p, nil
	}
	return "", ErrConfigFileNotFound
}

func (finder *configFinder) Finds(wd, configFilePath string) []string {
	if configFilePath == "" {
		return findconfig.Finds(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
	}
	return append([]string{configFilePath}, findconfig.Finds(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")...)
}
