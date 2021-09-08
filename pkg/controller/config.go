package controller

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type Package struct {
	Name     string `validate:"required"`
	Registry string `validate:"required"`
	Version  string `validate:"required"`
}

type Config struct {
	Packages       []*Package       `validate:"dive"`
	InlineRegistry *RegistryContent `yaml:"inline_registry"`
	Registries     Registries       `validate:"dive"`
}

type (
	PackageInfos []PackageInfo
	Registries   []Registry
)

const (
	pkgInfoTypeGitHubRelease = "github_release"
	pkgInfoTypeHTTP          = "http"
)

var (
	errPkgInfoNameIsDuplicated = errors.New("the package info name must be unique in the same registry")
	errInvalidType             = errors.New("type is invalid")
)

func (pkgInfos *PackageInfos) ToMap() (map[string]PackageInfo, error) {
	m := make(map[string]PackageInfo, len(*pkgInfos))
	for _, pkgInfo := range *pkgInfos {
		if _, ok := m[pkgInfo.GetName()]; ok {
			return nil, logerr.WithFields(errPkgInfoNameIsDuplicated, logrus.Fields{ //nolint:wrapcheck
				"package_info_name": pkgInfo.GetName(),
			})
		}
		m[pkgInfo.GetName()] = pkgInfo
	}
	return m, nil
}

func (pkgInfos *PackageInfos) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var arr []mergedPackageInfo
	if err := unmarshal(&arr); err != nil {
		return err
	}
	list := make([]PackageInfo, len(arr))
	for i, p := range arr {
		pkgInfo, err := p.GetPackageInfo()
		if err != nil {
			return err
		}
		list[i] = pkgInfo
	}
	*pkgInfos = list
	return nil
}

func (registries *Registries) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var arr []mergedRegistry
	if err := unmarshal(&arr); err != nil {
		return err
	}
	list := make([]Registry, len(arr))
	for i, p := range arr {
		registry, err := p.GetRegistry()
		if err != nil {
			return err
		}
		list[i] = registry
	}
	*registries = list
	return nil
}

type PackageInfo interface {
	GetName() string
	GetType() string
	GetArchiveType() string
	GetFiles() []*File
	GetFileSrc(pkg *Package, file *File) (string, error)
	GetPkgPath(rootDir string, pkg *Package) (string, error)
	RenderAsset(pkg *Package) (string, error)
	GetLink() string
	GetDescription() string
}

type mergedPackageInfo struct {
	Name        string
	Type        string
	RepoOwner   string `yaml:"repo_owner"`
	RepoName    string `yaml:"repo_name"`
	Asset       *text.Template
	ArchiveType string `yaml:"archive_type"`
	Files       []*File
	URL         *text.Template
	Description string
	Link        string
}

func (pkgInfo *mergedPackageInfo) GetPackageInfo() (PackageInfo, error) {
	switch pkgInfo.Type {
	case pkgInfoTypeGitHubRelease:
		return &GitHubReleasePackageInfo{
			Name:        pkgInfo.Name,
			RepoOwner:   pkgInfo.RepoOwner,
			RepoName:    pkgInfo.RepoName,
			Asset:       pkgInfo.Asset,
			ArchiveType: pkgInfo.ArchiveType,
			Files:       pkgInfo.Files,
			Link:        pkgInfo.Link,
			Description: pkgInfo.Description,
		}, nil
	case pkgInfoTypeHTTP:
		return &HTTPPackageInfo{
			Name:        pkgInfo.Name,
			ArchiveType: pkgInfo.ArchiveType,
			URL:         pkgInfo.URL,
			Files:       pkgInfo.Files,
			Link:        pkgInfo.Link,
			Description: pkgInfo.Description,
		}, nil
	default:
		return nil, logerr.WithFields(errInvalidType, logrus.Fields{ //nolint:wrapcheck
			"package_name": pkgInfo.Name,
			"package_type": pkgInfo.Type,
		})
	}
}

type File struct {
	Name string `validate:"required"`
	Src  *text.Template
}

func (file *File) RenderSrc(pkg *Package, pkgInfo PackageInfo) (string, error) {
	return file.Src.Execute(map[string]interface{}{ //nolint:wrapcheck
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
		"File":        file,
	})
}

type Param struct {
	ConfigFilePath string
	LogLevel       string
	OnlyLink       bool
	IsTest         bool
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
	FindGlobal(rootDir string) string
}

type configFinder struct{}

func (finder *configFinder) Find(wd string) string {
	return findconfig.Find(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
}

func (finder *configFinder) FindGlobal(rootDir string) string {
	for _, file := range []string{"aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml"} {
		cfgFilePath := filepath.Join(rootDir, "global", file)
		if _, err := os.Stat(cfgFilePath); err == nil {
			return cfgFilePath
		}
	}
	return ""
}

type ConfigReader interface {
	Read(p string) (io.ReadCloser, error)
}

type configReader struct{}

func (reader *configReader) Read(p string) (io.ReadCloser, error) {
	return os.Open(p) //nolint:wrapcheck
}
