package controller

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/suzuki-shunsuke/go-findconfig/findconfig"
	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Packages       []*Package   `validate:"dive"`
	InlineRegistry PackageInfos `yaml:"inline_registry" validate:"dive"`
}

type PackageInfos []PackageInfo

const (
	pkgInfoTypeGitHubRelease = "github_release"
	pkgInfoTypeHTTP          = "http"
)

func (pkgInfos *PackageInfos) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var arr []mergedPackageInfo
	if err := unmarshal(&arr); err != nil {
		return err
	}
	list := make([]PackageInfo, len(arr))
	for i, p := range arr {
		switch p.Type {
		case pkgInfoTypeGitHubRelease:
			pkgInfo := &GitHubReleasePackageInfo{
				Name:        p.Name,
				RepoOwner:   p.RepoOwner,
				RepoName:    p.RepoName,
				Asset:       p.Asset,
				ArchiveType: p.ArchiveType,
				Files:       p.Files,
			}
			list[i] = pkgInfo
		case pkgInfoTypeHTTP:
			pkgInfo := &HTTPPackageInfo{
				Name:        p.Name,
				ArchiveType: p.ArchiveType,
				URL:         p.URL,
				Files:       p.Files,
			}
			list[i] = pkgInfo
		default:
			return errors.New("type is invalid")
		}
	}
	*pkgInfos = list
	return nil
}

type Package struct {
	Name     string `validate:"required"`
	Registry string `validate:"required"`
	Version  string `validate:"required"`
}

type PackageInfo interface {
	GetName() string
	GetType() string
	GetArchiveType() string
	GetFiles() []*File
	GetFileSrc(pkg *Package, file *File) (string, error)
	GetPkgPath(rootDir string, pkg *Package) (string, error)
	RenderAsset(pkg *Package) (string, error)
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
}

type GitHubReleasePackageInfo struct {
	Name        string  `validate:"required"`
	ArchiveType string  `yaml:"archive_type"`
	Files       []*File `validate:"required,dive"`

	RepoOwner string         `yaml:"repo_owner" validate:"required"`
	RepoName  string         `yaml:"repo_name" validate:"required"`
	Asset     *text.Template `validate:"required"`
}

func (pkgInfo *GitHubReleasePackageInfo) GetName() string {
	return pkgInfo.Name
}

func (pkgInfo *GitHubReleasePackageInfo) GetType() string {
	return pkgInfoTypeGitHubRelease
}

func (pkgInfo *GitHubReleasePackageInfo) GetArchiveType() string {
	return pkgInfo.ArchiveType
}

func (pkgInfo *GitHubReleasePackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
}

func (pkgInfo *GitHubReleasePackageInfo) GetFiles() []*File {
	return pkgInfo.Files
}

func (pkgInfo *GitHubReleasePackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if isUnarchived(pkgInfo.GetArchiveType(), assetName) {
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

func (pkgInfo *GitHubReleasePackageInfo) RenderAsset(pkg *Package) (string, error) {
	return pkgInfo.Asset.Execute(map[string]interface{}{ //nolint:wrapcheck
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
	})
}

type HTTPPackageInfo struct {
	Name        string  `validate:"required"`
	ArchiveType string  `yaml:"archive_type"`
	Files       []*File `validate:"required,dive"`

	URL *text.Template `validate:"required"`
}

func (pkgInfo *HTTPPackageInfo) GetName() string {
	return pkgInfo.Name
}

func (pkgInfo *HTTPPackageInfo) GetType() string {
	return pkgInfoTypeHTTP
}

func (pkgInfo *HTTPPackageInfo) GetArchiveType() string {
	return pkgInfo.ArchiveType
}

func (pkgInfo *HTTPPackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	uS, err := pkgInfo.URL.Execute(map[string]interface{}{
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
	})
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}
	u, err := url.Parse(uS) // TODO
	if err != nil {
		return "", fmt.Errorf("parse the URL: %w", err)
	}
	return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), u.Host, u.Path), nil
}

func (pkgInfo *HTTPPackageInfo) GetFiles() []*File {
	return pkgInfo.Files
}

func (pkgInfo *HTTPPackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if isUnarchived(pkgInfo.GetArchiveType(), assetName) {
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

func (pkgInfo *HTTPPackageInfo) RenderURL(pkg *Package) (string, error) {
	uS, err := pkgInfo.URL.Execute(map[string]interface{}{
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
	})
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}
	return uS, nil
}

func (pkgInfo *HTTPPackageInfo) RenderAsset(pkg *Package) (string, error) {
	uS, err := pkgInfo.RenderURL(pkg)
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}
	u, err := url.Parse(uS) // TODO
	if err != nil {
		return "", fmt.Errorf("parse the URL: %w", err)
	}
	return filepath.Base(u.Path), nil
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
