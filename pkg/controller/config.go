package controller

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type Package struct {
	Name     string `validate:"required"`
	Registry string `validate:"required" yaml:",omitempty"`
	Version  string `validate:"required" yaml:",omitempty"`
	Import   string `yaml:",omitempty"`
}

func (pkg *Package) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Package
	a := alias(*pkg)
	if err := unmarshal(&a); err != nil {
		return err
	}
	name, version := parseNameWithVersion(a.Name)
	if name != "" {
		a.Name = name
	}
	if version != "" {
		a.Version = version
	}
	*pkg = Package(a)
	if pkg.Registry == "" {
		pkg.Registry = registryTypeStandard
	}

	return nil
}

func parseNameWithVersion(name string) (string, string) {
	idx := strings.Index(name, "@")
	if idx == -1 {
		return name, ""
	}
	return name[:idx], name[idx+1:]
}

type Config struct {
	Packages       []*Package       `validate:"dive"`
	InlineRegistry *RegistryContent `yaml:"inline_registry"`
	Registries     Registries       `validate:"dive"`
}

type (
	PackageInfos []*PackageInfo
	Registries   []*Registry
)

const (
	pkgInfoTypeGitHubRelease = "github_release"
	pkgInfoTypeGitHubContent = "github_content"
	pkgInfoTypeGitHubArchive = "github_archive"
	pkgInfoTypeHTTP          = "http"
)

func (pkgInfos *PackageInfos) ToMap() (map[string]*PackageInfo, error) {
	m := make(map[string]*PackageInfo, len(*pkgInfos))
	for _, pkgInfo := range *pkgInfos {
		pkgInfo := pkgInfo
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

type FormatOverride struct {
	GOOS   string
	Format string `yaml:"format"`
}

type File struct {
	Name string `validate:"required"`
	Src  *Template
}

func getArch(rosetta2 bool, replacements map[string]string) string {
	if rosetta2 && runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		// Rosetta 2
		return replace("amd64", replacements)
	}
	return replace(runtime.GOARCH, replacements)
}

func (file *File) RenderSrc(pkg *Package, pkgInfo *PackageInfo) (string, error) {
	return file.Src.Execute(map[string]interface{}{
		"Version":  pkg.Version,
		"GOOS":     runtime.GOOS,
		"GOARCH":   runtime.GOARCH,
		"OS":       replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":     getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements()),
		"Format":   pkgInfo.GetFormat(),
		"FileName": file.Name,
	})
}

type Param struct {
	ConfigFilePath string
	LogLevel       string
	OnlyLink       bool
	IsTest         bool
	All            bool
	File           string
	GlobalConfigs  []string
	AQUAVersion    string
}

func (ctrl *Controller) getConfigFilePath(wd, configFilePath string) string {
	if configFilePath != "" {
		return configFilePath
	}
	return ctrl.ConfigFinder.Find(wd)
}

func (ctrl *Controller) getConfigFilePaths(wd, configFilePath string) []string {
	if configFilePath == "" {
		return ctrl.ConfigFinder.Finds(wd)
	}
	return append([]string{configFilePath}, ctrl.ConfigFinder.Finds(wd)...)
}

func (ctrl *Controller) readImports(configFilePath string, cfg *Config) error {
	pkgs := []*Package{}
	for _, pkg := range cfg.Packages {
		if pkg.Import == "" {
			pkgs = append(pkgs, pkg)
			continue
		}
		p := filepath.Join(filepath.Dir(configFilePath), pkg.Import)
		filePaths, err := filepath.Glob(p)
		if err != nil {
			return fmt.Errorf("read files with glob pattern (%s): %w", p, err)
		}
		sort.Strings(filePaths)
		for _, filePath := range filePaths {
			subCfg := &Config{}
			if err := ctrl.readConfig(filePath, subCfg); err != nil {
				return err
			}
			pkgs = append(pkgs, subCfg.Packages...)
		}
	}
	cfg.Packages = pkgs
	return nil
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
	if err := ctrl.readImports(configFilePath, cfg); err != nil {
		return fmt.Errorf("read imports (%s): %w", configFilePath, err)
	}
	return nil
}

type ConfigFinder interface {
	Find(wd string) string
	Finds(wd string) []string
	FindGlobal(rootDir string) string
}

type configFinder struct{}

func (finder *configFinder) Find(wd string) string {
	return findconfig.Find(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
}

func (finder *configFinder) Finds(wd string) []string {
	return findconfig.Finds(wd, findconfig.Exist, "aqua.yaml", "aqua.yml", ".aqua.yaml", ".aqua.yml")
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
