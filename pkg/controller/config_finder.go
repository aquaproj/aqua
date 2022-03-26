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

func (ctrl *Controller) getConfigFilePaths(wd, configFilePath string) []string {
	if configFilePath == "" {
		return ctrl.ConfigFinder.Finds(wd)
	}
	return append([]string{configFilePath}, ctrl.ConfigFinder.Finds(wd)...)
}

type configFinder struct{}

func (finder *configFinder) Find(wd string) string {
	return findconfig.Find(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
}

func (finder *configFinder) Finds(wd string) []string {
	return findconfig.Finds(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
}

func (finder *configFinder) FindFirstConfig(wd string) (string, error) {
	if cfgFilePath := finder.Find(wd); cfgFilePath != "" {
		return cfgFilePath, nil
	}
	for _, p := range getGlobalConfigFilePaths() {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		return p, nil
	}
	return "", errConfigFileNotFound
}
