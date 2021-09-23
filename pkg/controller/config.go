package controller

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
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

func (pkgInfos *PackageInfos) ToMap() (map[string]PackageInfo, error) {
	m := make(map[string]PackageInfo, len(*pkgInfos))
	for _, pkgInfo := range *pkgInfos {
		name := pkgInfo.GetName()
		if _, ok := m[name]; ok {
			return nil, logerr.WithFields(errPkgNameMustBeUniqueInRegistry, logrus.Fields{ //nolint:wrapcheck
				"package_name": name,
			})
		}
		m[name] = pkgInfo
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
	GetFormat() string
	GetFiles() []*File
	GetFileSrc(pkg *Package, file *File) (string, error)
	GetPkgPath(rootDir string, pkg *Package) (string, error)
	RenderAsset(pkg *Package) (string, error)
	GetLink() string
	GetDescription() string
	GetReplacements() map[string]string
	SetVersion(v string) (PackageInfo, error)
}

type mergedPackageInfo struct {
	Name               string
	Type               string
	RepoOwner          string `yaml:"repo_owner"`
	RepoName           string `yaml:"repo_name"`
	Asset              *Template
	Format             string
	Files              []*File
	URL                *Template
	Description        string
	Link               string
	Replacements       map[string]string
	FormatOverrides    []*FormatOverride        `yaml:"format_overrides"`
	VersionConstraints *VersionConstraints      `yaml:"version_constraint"`
	VersionOverrides   []*mergedVersionOverride `yaml:"version_overrides"`
}

type mergedVersionOverride struct {
	VersionConstraints *VersionConstraints `yaml:"version_constraint"`
	Asset              *Template
	Files              []*File `validate:"dive"`
	Format             string
	FormatOverrides    []*FormatOverride `yaml:"format_overrides"`
	Replacements       map[string]string
}

type FormatOverride struct {
	GOOS   string
	Format string `yaml:"format"`
}

func (pkgInfo *mergedPackageInfo) GetPackageInfo() (PackageInfo, error) {
	switch pkgInfo.Type {
	case pkgInfoTypeGitHubRelease:
		var versionOverrides []*GitHubReleaseVersionOverride
		if pkgInfo.VersionOverrides != nil {
			versionOverrides = make([]*GitHubReleaseVersionOverride, len(pkgInfo.VersionOverrides))
			for i, vo := range pkgInfo.VersionOverrides {
				versionOverrides[i] = &GitHubReleaseVersionOverride{
					VersionConstraints: vo.VersionConstraints,
					Asset:              vo.Asset,
					Files:              vo.Files,
					Format:             vo.Format,
					FormatOverrides:    vo.FormatOverrides,
					Replacements:       vo.Replacements,
				}
			}
		}
		return &GitHubReleasePackageInfo{
			Name:               pkgInfo.Name,
			RepoOwner:          pkgInfo.RepoOwner,
			RepoName:           pkgInfo.RepoName,
			Asset:              pkgInfo.Asset,
			Format:             pkgInfo.Format,
			FormatOverrides:    pkgInfo.FormatOverrides,
			Files:              pkgInfo.Files,
			Link:               pkgInfo.Link,
			Description:        pkgInfo.Description,
			Replacements:       pkgInfo.Replacements,
			VersionConstraints: pkgInfo.VersionConstraints,
			VersionOverrides:   versionOverrides,
		}, nil
	case pkgInfoTypeHTTP:
		return &HTTPPackageInfo{
			Name:            pkgInfo.Name,
			Format:          pkgInfo.Format,
			FormatOverrides: pkgInfo.FormatOverrides,
			URL:             pkgInfo.URL,
			Files:           pkgInfo.Files,
			Link:            pkgInfo.Link,
			Description:     pkgInfo.Description,
			Replacements:    pkgInfo.Replacements,
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
	Src  *Template
}

func (file *File) RenderSrc(pkg *Package, pkgInfo PackageInfo) (string, error) {
	return file.Src.Execute(map[string]interface{}{
		"Version":  pkg.Version,
		"GOOS":     runtime.GOOS,
		"GOARCH":   runtime.GOARCH,
		"OS":       replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":     replace(runtime.GOARCH, pkgInfo.GetReplacements()),
		"Format":   pkgInfo.GetFormat(),
		"FileName": file.Name,
	})
}

type Param struct {
	ConfigFilePath string
	LogLevel       string
	OnlyLink       bool
	IsTest         bool
	File           string
	AQUAVersion    string
}

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
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
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
