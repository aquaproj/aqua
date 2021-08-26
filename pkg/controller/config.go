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
	Packages         []*Package     `validate:"dive"`
	InlineRepository []*PackageInfo `yaml:"inline_repository" validate:"dive"`
	BinDir           string         `yaml:"bin_dir"`
}

type Package struct {
	Name       string `validate:"required"`
	Repository string `validate:"required"`
	Version    string `validate:"required"`
}

type PackageInfo struct {
	Name        string         `validate:"required"`
	Type        string         `validate:"required"`
	RepoOwner   string         `yaml:"repo_owner" validate:"required"`
	RepoName    string         `yaml:"repo_name" validate:"required"`
	Artifact    *text.Template `validate:"required"`
	ArchiveType string         `yaml:"archive_type"`
	Files       []*File        `validate:"required,dive"`
}

type File struct {
	Name string `validate:"required"`
	Src  *text.Template
}

type Param struct {
	ConfigFilePath string
	LogLevel       string
}

var errConfigFileNotFound = errors.New("configuration file isn't found")

func (ctrl *Controller) getConfigFilePath(wd, configFilePath string) string {
	if configFilePath != "" {
		return configFilePath
	}
	return ctrl.ConfigFinder.Find(wd)
}

func (ctrl *Controller) readConfig(configFilePath string, cfg *Config) error {
	file, err := ctrl.ConfigReader.Read(configFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	if err := decoder.Decode(&cfg); err != nil {
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
