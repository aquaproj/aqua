package controller

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Packages         []*Package
	InlineRepository []*PackageInfo `yaml:"inline_repository"`
	BinDir           string         `yaml:"bin_dir"`
}

type Package struct {
	Name       string
	Repository string
	Version    string
}

type PackageInfo struct {
	Name        string
	Type        string
	Repo        string
	Artifact    *text.Template
	ArchiveType string `yaml:"archive_type"`
	Files       []*File
}

type File struct {
	Name string
	Src  string
}

type Param struct {
	ConfigFilePath string
	LogLevel       string
}

var errConfigFileNotFound = errors.New("configuration file isn't found")

func (ctrl *Controller) readConfig(wd, configFilePath string, cfg *Config) error {
	if configFilePath == "" {
		p := ctrl.ConfigFinder.Find(wd)
		if p == "" {
			return errConfigFileNotFound
		}
		configFilePath = p
	}
	file, err := ctrl.ConfigReader.Read(configFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer file.Close()
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return fmt.Errorf("parse a configuration file as YAML %s: %w", configFilePath, err)
	}
	return nil
}

type ConfigFinder interface {
	Find(wd string) string
}

type configFinder struct{}

func (finder *configFinder) Find(wd string) string {
	return findconfig.Find(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
}

type ConfigReader interface {
	Read(p string) (io.ReadCloser, error)
}

type configReader struct{}

func (reader *configReader) Read(p string) (io.ReadCloser, error) {
	return os.Open(p) //nolint:wrapcheck
}
