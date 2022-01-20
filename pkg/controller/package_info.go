package controller

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
)

type PackageInfo struct {
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
	FormatOverrides    []*FormatOverride   `yaml:"format_overrides"`
	VersionConstraints *VersionConstraints `yaml:"version_constraint"`
	VersionOverrides   []*PackageInfo      `yaml:"version_overrides"`
	SupportedIf        *PackageCondition   `yaml:"supported_if"`
	VersionFilter      *VersionFilter      `yaml:"version_filter"`
	Rosetta2           *bool
}

func (pkgInfo *PackageInfo) GetRosetta2() bool {
	return pkgInfo.Rosetta2 != nil && *pkgInfo.Rosetta2
}

func (pkgInfo *PackageInfo) HasRepo() bool {
	return pkgInfo.RepoOwner != "" && pkgInfo.RepoName != ""
}

func (pkgInfo *PackageInfo) GetName() string {
	if pkgInfo.Name != "" {
		return pkgInfo.Name
	}
	if pkgInfo.HasRepo() {
		return pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
	}
	return ""
}

func (pkgInfo *PackageInfo) GetLink() string {
	if pkgInfo.Link != "" {
		return pkgInfo.Link
	}
	if pkgInfo.HasRepo() {
		return "https://github.com/" + pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
	}
	return ""
}

func (pkgInfo *PackageInfo) GetFormat() string {
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

func (pkgInfo *PackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if isUnarchived(pkgInfo.GetFormat(), assetName) {
		return filepath.Base(assetName), nil
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

func (pkgInfo *PackageInfo) GetDescription() string {
	return pkgInfo.Description
}

func (pkgInfo *PackageInfo) GetType() string {
	return pkgInfo.Type
}

func (pkgInfo *PackageInfo) SetVersion(v string) (*PackageInfo, error) {
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

func (pkgInfo *PackageInfo) override(child *PackageInfo) { //nolint:cyclop
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
	if child.SupportedIf != nil {
		pkgInfo.SupportedIf = child.SupportedIf
	}
	if child.Rosetta2 != nil {
		pkgInfo.Rosetta2 = child.Rosetta2
	}
	if child.VersionFilter != nil {
		pkgInfo.VersionFilter = child.VersionFilter
	}
}

func (pkgInfo *PackageInfo) GetReplacements() map[string]string {
	return pkgInfo.Replacements
}

func (pkgInfo *PackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
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
			"Arch":    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements()),
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

func (pkgInfo *PackageInfo) validate() error { //nolint:cyclop
	if name := pkgInfo.GetName(); name == "" {
		return errPkgNameIsRequired
	}
	switch pkgInfo.Type {
	case pkgInfoTypeGitHubArchive:
		if !pkgInfo.HasRepo() {
			return errRepoRequired
		}
		return nil
	case pkgInfoTypeGitHubContent:
		if !pkgInfo.HasRepo() {
			return errRepoRequired
		}
		if pkgInfo.Path == nil {
			return errGitHubContentRequirePath
		}
		return nil
	case pkgInfoTypeGitHubRelease:
		if !pkgInfo.HasRepo() {
			return errRepoRequired
		}
		if pkgInfo.Asset == nil {
			return errAssetRequired
		}
		return nil
	case pkgInfoTypeHTTP:
		if pkgInfo.URL == nil {
			return errURLRequired
		}
		return nil
	}
	return errInvalidPackageType
}

func (pkgInfo *PackageInfo) RenderAsset(pkg *Package) (string, error) {
	switch pkgInfo.Type {
	case pkgInfoTypeGitHubArchive:
		return "", nil
	case pkgInfoTypeGitHubContent:
		return pkgInfo.Path.Execute(map[string]interface{}{
			"Version": pkg.Version,
			"GOOS":    runtime.GOOS,
			"GOARCH":  runtime.GOARCH,
			"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
			"Arch":    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements()),
			"Format":  pkgInfo.GetFormat(),
		})
	case pkgInfoTypeGitHubRelease:
		return pkgInfo.Asset.Execute(map[string]interface{}{
			"Version": pkg.Version,
			"GOOS":    runtime.GOOS,
			"GOARCH":  runtime.GOARCH,
			"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
			"Arch":    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements()),
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

func (pkgInfo *PackageInfo) renderURL(pkg *Package) (string, error) {
	uS, err := pkgInfo.URL.Execute(map[string]interface{}{
		"Version": pkg.Version,
		"GOOS":    runtime.GOOS,
		"GOARCH":  runtime.GOARCH,
		"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements()),
		"Format":  pkgInfo.GetFormat(),
	})
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}
	return uS, nil
}

func (pkgInfo *PackageInfo) GetFiles() []*File {
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
