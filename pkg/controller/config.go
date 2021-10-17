package controller

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
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
		pkg.Registry = "standard"
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
	PackageInfos []*MergedPackageInfo
	Registries   []Registry
)

const (
	pkgInfoTypeGitHubRelease = "github_release"
	pkgInfoTypeGitHubContent = "github_content"
	pkgInfoTypeGitHubArchive = "github_archive"
	pkgInfoTypeHTTP          = "http"
)

func (pkgInfos *PackageInfos) ToMap() (map[string]*MergedPackageInfo, error) {
	m := make(map[string]*MergedPackageInfo, len(*pkgInfos))
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

type MergedPackageInfo struct {
	Name               string
	Type               string `validate:"required"`
	RepoOwner          string `yaml:"repo_owner"`
	RepoName           string `yaml:"repo_name"`
	Asset              *Template
	Path               *Template
	Format             string
	Files              []*File
	URL                *Template
	Description        string
	Link               string
	Replacements       map[string]string
	FormatOverrides    []*FormatOverride    `yaml:"format_overrides"`
	VersionConstraints *VersionConstraints  `yaml:"version_constraint"`
	VersionOverrides   []*MergedPackageInfo `yaml:"version_overrides"`
}

func (pkgInfo *MergedPackageInfo) HasRepo() bool {
	switch pkgInfo.Type {
	case pkgInfoTypeGitHubRelease, pkgInfoTypeGitHubArchive, pkgInfoTypeGitHubContent:
		return true
	}
	return false
}

func (pkgInfo *MergedPackageInfo) GetName() string {
	if pkgInfo.Name != "" {
		return pkgInfo.Name
	}
	if pkgInfo.HasRepo() {
		return pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
	}
	return ""
}

func (pkgInfo *MergedPackageInfo) GetLink() string {
	if pkgInfo.Link != "" {
		return pkgInfo.Link
	}
	if pkgInfo.HasRepo() {
		return "https://github.com/" + pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
	}
	return ""
}

func (pkgInfo *MergedPackageInfo) GetFormat() string {
	if pkgInfo.Type == pkgInfoTypeGitHubArchive {
		return "tar.gz"
	}
	for _, arcTypeOverride := range pkgInfo.FormatOverrides {
		if arcTypeOverride.GOOS == runtime.GOOS {
			return arcTypeOverride.Format
		}
	}
	return pkgInfo.Format
}

func (pkgInfo *MergedPackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if isUnarchived(pkgInfo.GetFormat(), assetName) {
		return assetName, nil
	}
	if file.Src == nil {
		return file.Name, nil
	}
	src, err := file.RenderSrc(pkg, pkgInfo)
	if err != nil {
		return "", fmt.Errorf("render the template file.src: %w", err)
	}
	return src, nil
}

func (pkgInfo *MergedPackageInfo) GetDescription() string {
	return pkgInfo.Description
}

func (pkgInfo *MergedPackageInfo) GetType() string {
	return pkgInfo.Type
}

func (pkgInfo *MergedPackageInfo) SetVersion(v string) (*MergedPackageInfo, error) {
	if pkgInfo.VersionConstraints == nil {
		return pkgInfo, nil
	}
	a, err := pkgInfo.VersionConstraints.Check(v)
	if err != nil {
		return nil, err
	}
	if a {
		return pkgInfo, nil
	}
	for _, vo := range pkgInfo.VersionOverrides {
		a, err := vo.VersionConstraints.Check(v)
		if err != nil {
			return nil, err
		}
		if a {
			pkgInfo.override(vo)
			return pkgInfo, nil
		}
	}
	return pkgInfo, nil
}

func (pkgInfo *MergedPackageInfo) override(child *MergedPackageInfo) { //nolint:cyclop
	if child.Type != "" {
		pkgInfo.Type = child.Type
	}
	if child.RepoOwner != "" {
		pkgInfo.RepoOwner = child.RepoOwner
	}
	if child.RepoName != "" {
		pkgInfo.RepoName = child.RepoName
	}
	if child.Asset != nil {
		pkgInfo.Asset = child.Asset
	}
	if child.Path != nil {
		pkgInfo.Path = child.Path
	}
	if child.Format != "" {
		pkgInfo.Format = child.Format
	}
	if child.Files != nil {
		pkgInfo.Files = child.Files
	}
	if child.URL != nil {
		pkgInfo.URL = child.URL
	}
	if child.Replacements != nil {
		pkgInfo.Replacements = child.Replacements
	}
	if child.FormatOverrides != nil {
		pkgInfo.FormatOverrides = child.FormatOverrides
	}
}

func (pkgInfo *MergedPackageInfo) GetReplacements() map[string]string {
	return pkgInfo.Replacements
}

func (pkgInfo *MergedPackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	switch pkgInfo.Type {
	case pkgInfoTypeGitHubArchive:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case pkgInfoTypeGitHubContent, pkgInfoTypeGitHubRelease:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case pkgInfoTypeHTTP:
		uS, err := pkgInfo.URL.Execute(map[string]interface{}{
			"Version": pkg.Version,
			"GOOS":    runtime.GOOS,
			"GOARCH":  runtime.GOARCH,
			"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
			"Arch":    replace(runtime.GOARCH, pkgInfo.GetReplacements()),
			"Format":  pkgInfo.GetFormat(),
		})
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), u.Host, u.Path), nil
	}
	return "", nil
}

func (pkgInfo *MergedPackageInfo) RenderAsset(pkg *Package) (string, error) {
	switch pkgInfo.Type {
	case pkgInfoTypeGitHubArchive:
		return "", nil
	case pkgInfoTypeGitHubContent:
		return pkgInfo.Path.Execute(map[string]interface{}{
			"Version": pkg.Version,
			"GOOS":    runtime.GOOS,
			"GOARCH":  runtime.GOARCH,
			"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
			"Arch":    replace(runtime.GOARCH, pkgInfo.GetReplacements()),
			"Format":  pkgInfo.GetFormat(),
		})
	case pkgInfoTypeGitHubRelease:
		return pkgInfo.Asset.Execute(map[string]interface{}{
			"Version": pkg.Version,
			"GOOS":    runtime.GOOS,
			"GOARCH":  runtime.GOARCH,
			"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
			"Arch":    replace(runtime.GOARCH, pkgInfo.GetReplacements()),
			"Format":  pkgInfo.GetFormat(),
		})
	case pkgInfoTypeHTTP:
		uS, err := pkgInfo.renderURL(pkg)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return filepath.Base(u.Path), nil
	}
	return "", nil
}

func (pkgInfo *MergedPackageInfo) renderURL(pkg *Package) (string, error) {
	uS, err := pkgInfo.URL.Execute(map[string]interface{}{
		"Version": pkg.Version,
		"GOOS":    runtime.GOOS,
		"GOARCH":  runtime.GOARCH,
		"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":    replace(runtime.GOARCH, pkgInfo.GetReplacements()),
		"Format":  pkgInfo.GetFormat(),
	})
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}
	return uS, nil
}

func (pkgInfo *MergedPackageInfo) GetFiles() []*File {
	if len(pkgInfo.Files) != 0 {
		return pkgInfo.Files
	}
	if pkgInfo.HasRepo() {
		return []*File{
			{
				Name: pkgInfo.RepoName,
			},
		}
	}
	return pkgInfo.Files
}

type FormatOverride struct {
	GOOS   string
	Format string `yaml:"format"`
}

type File struct {
	Name string `validate:"required"`
	Src  *Template
}

func (file *File) RenderSrc(pkg *Package, pkgInfo *MergedPackageInfo) (string, error) {
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
