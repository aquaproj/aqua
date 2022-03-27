package controller

import "github.com/aquaproj/aqua/pkg/config"

type ConfigFinder interface {
	Find(wd, configFilePath string) (string, error)
	Finds(wd, configFilePath string) []string
	GetGlobalConfigFilePaths() []string
}

type ConfigReader interface {
	Read(configFilePath string, cfg *config.Config) error
}
