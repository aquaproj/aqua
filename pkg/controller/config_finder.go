package controller

import (
	"os"
	"strings"

	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
)

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

func (finder *configFinder) Find(wd, configFilePath string) (string, error) {
	if configFilePath != "" {
		return configFilePath, nil
	}
	configFilePath = findconfig.Find(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
	if configFilePath != "" {
		return configFilePath, nil
	}
	for _, p := range getGlobalConfigFilePaths() {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		return p, nil
	}
	return "", errConfigFileNotFound
}

func (finder *configFinder) Finds(wd, configFilePath string) []string {
	if configFilePath == "" {
		return findconfig.Finds(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
	}
	return append([]string{configFilePath}, findconfig.Finds(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")...)
}
