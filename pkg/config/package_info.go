package config

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/aquaproj/aqua/pkg/unarchive"
	constraint "github.com/aquaproj/aqua/pkg/version-constraint"
)

type PackageInfo struct {
	Name               string                         `json:"name,omitempty"`
	Type               string                         `validate:"required" json:"type" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http"`
	RepoOwner          string                         `yaml:"repo_owner" json:"repo_owner,omitempty"`
	RepoName           string                         `yaml:"repo_name" json:"repo_name,omitempty"`
	Asset              *template.Template             `json:"asset,omitempty"`
	Path               *template.Template             `json:"path,omitempty"`
	Format             string                         `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw"`
	Files              []*File                        `json:"files,omitempty"`
	URL                *template.Template             `json:"url,omitempty"`
	Description        string                         `json:"description,omitempty"`
	Link               string                         `json:"link,omitempty"`
	Replacements       map[string]string              `json:"replacements,omitempty"`
	Overrides          []*Override                    `json:"overrides,omitempty"`
	FormatOverrides    []*FormatOverride              `yaml:"format_overrides" json:"format_overrides,omitempty"`
	VersionConstraints *constraint.VersionConstraints `yaml:"version_constraint" json:"version_constraint,omitempty"`
	VersionOverrides   []*VersionOverride             `yaml:"version_overrides" json:"version_overrides,omitempty"`
	SupportedIf        *constraint.PackageCondition   `yaml:"supported_if" json:"supported_if,omitempty"`
	VersionFilter      *constraint.VersionFilter      `yaml:"version_filter" json:"version_filter,omitempty"`
	Rosetta2           *bool                          `json:"rosetta2,omitempty"`
	Aliases            []*Alias                       `json:"aliases,omitempty"`
}

type Alias struct {
	Name string `json:"name"`
}

type VersionOverride struct {
	Type               string                         `json:"type,omitempty" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http"`
	RepoOwner          string                         `yaml:"repo_owner" json:"repo_owner,omitempty"`
	RepoName           string                         `yaml:"repo_name" json:"repo_name,omitempty"`
	Asset              *template.Template             `json:"asset,omitempty"`
	Path               *template.Template             `json:"path,omitempty"`
	Format             string                         `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw"`
	Files              []*File                        `json:"files,omitempty"`
	URL                *template.Template             `json:"url,omitempty"`
	Replacements       map[string]string              `json:"replacements,omitempty"`
	Overrides          []*Override                    `json:"overrides,omitempty"`
	FormatOverrides    []*FormatOverride              `yaml:"format_overrides" json:"format_overrides,omitempty"`
	SupportedIf        *constraint.PackageCondition   `yaml:"supported_if" json:"supported_if,omitempty"`
	VersionConstraints *constraint.VersionConstraints `yaml:"version_constraint" json:"version_constraint,omitempty"`
	VersionFilter      *constraint.VersionFilter      `yaml:"version_filter" json:"version_filter,omitempty"`
	Rosetta2           *bool                          `json:"rosetta2,omitempty"`
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

func (pkgInfo *PackageInfo) GetFileSrc(pkg *Package, file *File, rt *runtime.Runtime) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg, rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if unarchive.IsUnarchived(pkgInfo.GetFormat(), assetName) {
		return filepath.Base(assetName), nil
	}
	if file.Src == nil {
		return file.Name, nil
	}
	src, err := file.RenderSrc(pkg, pkgInfo, rt)
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

func (pkgInfo *PackageInfo) GetPkgPath(rootDir string, pkg *Package, rt *runtime.Runtime) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg, rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
		uS, err := pkgInfo.renderTemplate(pkgInfo.URL, pkg, rt)
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

func (pkgInfo *PackageInfo) RenderAsset(pkg *Package, rt *runtime.Runtime) (string, error) {
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return "", nil
	case PkgInfoTypeGitHubContent:
		s, err := pkgInfo.renderTemplate(pkgInfo.Path, pkg, rt)
		if err != nil {
			return "", fmt.Errorf("render a package path: %w", err)
		}
		return s, nil
	case PkgInfoTypeGitHubRelease:
		return pkgInfo.renderTemplate(pkgInfo.Asset, pkg, rt)
	case PkgInfoTypeHTTP:
		uS, err := pkgInfo.renderTemplate(pkgInfo.URL, pkg, rt)
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

func (pkgInfo *PackageInfo) renderTemplate(tpl *template.Template, pkg *Package, rt *runtime.Runtime) (string, error) {
	uS, err := tpl.Execute(map[string]interface{}{
		"Version": pkg.Version,
		"GOOS":    rt.GOOS,
		"GOARCH":  rt.GOARCH,
		"OS":      replace(rt.GOOS, pkgInfo.GetReplacements()),
		"Arch":    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements(), rt),
		"Format":  pkgInfo.GetFormat(),
	})
	if err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return uS, nil
}

func (pkgInfo *PackageInfo) RenderURL(pkg *Package, rt *runtime.Runtime) (string, error) {
	return pkgInfo.renderTemplate(pkgInfo.URL, pkg, rt)
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
