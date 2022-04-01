package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"

	"github.com/aquaproj/aqua/pkg/template"
	"github.com/aquaproj/aqua/pkg/unarchive"
	constraint "github.com/aquaproj/aqua/pkg/version-constraint"
)

type PackageInfo struct {
	Name               string                         `json:"name"`
	Type               string                         `validate:"required" json:"type"`
	RepoOwner          string                         `yaml:"repo_owner" json:"repo_owner"`
	RepoName           string                         `yaml:"repo_name" json:"repo_name"`
	Asset              *template.Template             `json:"asset"`
	Path               *template.Template             `json:"path"`
	Format             string                         `json:"format"`
	Files              []*File                        `json:"files"`
	URL                *template.Template             `json:"url"`
	Description        string                         `json:"description"`
	Link               string                         `json:"link"`
	Replacements       map[string]string              `json:"replacements"`
	Overrides          []*Override                    `json:"overrides"`
	FormatOverrides    []*FormatOverride              `yaml:"format_overrides" json:"format_overrides"`
	VersionConstraints *constraint.VersionConstraints `yaml:"version_constraint" json:"version_constraint"`
	VersionOverrides   []*PackageInfo                 `yaml:"version_overrides" json:"version_overrides"`
	SupportedIf        *constraint.PackageCondition   `yaml:"supported_if" json:"supported_if"`
	VersionFilter      *constraint.VersionFilter      `yaml:"version_filter" json:"version_filter"`
	Rosetta2           *bool                          `json:"rosseta2"`
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
	if pkgInfo.Type == PkgInfoTypeGitHubArchive {
		return "tar.gz"
	}
	return pkgInfo.Format
}

func (pkgInfo *PackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if unarchive.IsUnarchived(pkgInfo.GetFormat(), assetName) {
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

func (pkgInfo *PackageInfo) GetReplacements() map[string]string {
	return pkgInfo.Replacements
}

func (pkgInfo *PackageInfo) GetAsset() *template.Template {
	return pkgInfo.Asset
}

func (pkgInfo *PackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
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

func (pkgInfo *PackageInfo) Validate() error { //nolint:cyclop
	if name := pkgInfo.GetName(); name == "" {
		return errPkgNameIsRequired
	}
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		if !pkgInfo.HasRepo() {
			return errRepoRequired
		}
		return nil
	case PkgInfoTypeGitHubContent:
		if !pkgInfo.HasRepo() {
			return errRepoRequired
		}
		if pkgInfo.Path == nil {
			return errGitHubContentRequirePath
		}
		return nil
	case PkgInfoTypeGitHubRelease:
		if !pkgInfo.HasRepo() {
			return errRepoRequired
		}
		if pkgInfo.Asset == nil {
			return errAssetRequired
		}
		return nil
	case PkgInfoTypeHTTP:
		if pkgInfo.URL == nil {
			return errURLRequired
		}
		return nil
	}
	return errInvalidPackageType
}

func (pkgInfo *PackageInfo) RenderAsset(pkg *Package) (string, error) {
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return "", nil
	case PkgInfoTypeGitHubContent:
		return pkgInfo.Path.Execute(map[string]interface{}{ //nolint:wrapcheck
			"Version": pkg.Version,
			"GOOS":    runtime.GOOS,
			"GOARCH":  runtime.GOARCH,
			"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
			"Arch":    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements()),
			"Format":  pkgInfo.GetFormat(),
		})
	case PkgInfoTypeGitHubRelease:
		return pkgInfo.Asset.Execute(map[string]interface{}{ //nolint:wrapcheck
			"Version": pkg.Version,
			"GOOS":    runtime.GOOS,
			"GOARCH":  runtime.GOARCH,
			"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
			"Arch":    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements()),
			"Format":  pkgInfo.GetFormat(),
		})
	case PkgInfoTypeHTTP:
		uS, err := pkgInfo.RenderURL(pkg)
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

func (pkgInfo *PackageInfo) RenderURL(pkg *Package) (string, error) {
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
