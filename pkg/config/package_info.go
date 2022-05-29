package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	texttemplate "text/template"

	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/aquaproj/aqua/pkg/unarchive"
)

type PackageInfo struct {
	Name               string             `json:"name,omitempty"`
	Type               string             `validate:"required" json:"type" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go"`
	RepoOwner          string             `yaml:"repo_owner" json:"repo_owner,omitempty"`
	RepoName           string             `yaml:"repo_name" json:"repo_name,omitempty"`
	Asset              *string            `json:"asset,omitempty"`
	Path               *string            `json:"path,omitempty"`
	Format             string             `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw"`
	Files              []*File            `json:"files,omitempty"`
	URL                *string            `json:"url,omitempty"`
	Description        string             `json:"description,omitempty"`
	Link               string             `json:"link,omitempty"`
	Replacements       map[string]string  `json:"replacements,omitempty"`
	Overrides          []*Override        `json:"overrides,omitempty"`
	FormatOverrides    []*FormatOverride  `yaml:"format_overrides" json:"format_overrides,omitempty"`
	VersionConstraints string             `yaml:"version_constraint" json:"version_constraint,omitempty"`
	VersionOverrides   []*VersionOverride `yaml:"version_overrides" json:"version_overrides,omitempty"`
	SupportedIf        *string            `yaml:"supported_if" json:"supported_if,omitempty"`
	VersionFilter      *string            `yaml:"version_filter" json:"version_filter,omitempty"`
	Rosetta2           *bool              `json:"rosetta2,omitempty"`
	Aliases            []*Alias           `json:"aliases,omitempty"`
	VersionSource      string             `json:"version_source,omitempty" yaml:"version_source"`
}

func (pkgInfo *PackageInfo) copy() *PackageInfo {
	pkg := &PackageInfo{
		Name:               pkgInfo.Name,
		Type:               pkgInfo.Type,
		RepoOwner:          pkgInfo.RepoOwner,
		RepoName:           pkgInfo.RepoName,
		Asset:              pkgInfo.Asset,
		Path:               pkgInfo.Path,
		Format:             pkgInfo.Format,
		Files:              pkgInfo.Files,
		URL:                pkgInfo.URL,
		Description:        pkgInfo.Description,
		Link:               pkgInfo.Link,
		Replacements:       pkgInfo.Replacements,
		Overrides:          pkgInfo.Overrides,
		FormatOverrides:    pkgInfo.FormatOverrides,
		VersionConstraints: pkgInfo.VersionConstraints,
		VersionOverrides:   pkgInfo.VersionOverrides,
		SupportedIf:        pkgInfo.SupportedIf,
		VersionFilter:      pkgInfo.VersionFilter,
		Rosetta2:           pkgInfo.Rosetta2,
		Aliases:            pkgInfo.Aliases,
		VersionSource:      pkgInfo.VersionSource,
	}
	return pkg
}

func (pkgInfo *PackageInfo) overrideVersion(child *VersionOverride) *PackageInfo { //nolint:cyclop
	pkg := pkgInfo.copy()
	if child.Type != "" {
		pkg.Type = child.Type
	}
	if child.RepoOwner != "" {
		pkg.RepoOwner = child.RepoOwner
	}
	if child.RepoName != "" {
		pkg.RepoName = child.RepoName
	}
	if child.Asset != nil {
		pkg.Asset = child.Asset
	}
	if child.Path != nil {
		pkg.Path = child.Path
	}
	if child.Format != "" {
		pkg.Format = child.Format
	}
	if child.Files != nil {
		pkg.Files = child.Files
	}
	if child.URL != nil {
		pkg.URL = child.URL
	}
	if child.Replacements != nil {
		pkg.Replacements = child.Replacements
	}
	if child.Overrides != nil {
		pkg.Overrides = child.Overrides
	}
	if child.FormatOverrides != nil {
		pkg.FormatOverrides = child.FormatOverrides
	}
	if child.SupportedIf != nil {
		pkg.SupportedIf = child.SupportedIf
	}
	if child.VersionFilter != nil {
		pkg.VersionFilter = child.VersionFilter
	}
	if child.Rosetta2 != nil {
		pkg.Rosetta2 = child.Rosetta2
	}
	if child.VersionSource != "" {
		pkg.VersionSource = child.VersionSource
	}
	return pkg
}

func (pkgInfo *PackageInfo) override(rt *runtime.Runtime) { //nolint:cyclop
	for _, fo := range pkgInfo.FormatOverrides {
		if fo.GOOS == rt.GOOS {
			pkgInfo.Format = fo.Format
			break
		}
	}

	ov := pkgInfo.getOverride(rt)
	if ov == nil {
		return
	}

	if pkgInfo.Replacements == nil {
		pkgInfo.Replacements = ov.Replacements
	} else {
		replacements := make(map[string]string, len(pkgInfo.Replacements))
		for k, v := range pkgInfo.Replacements {
			replacements[k] = v
		}
		for k, v := range ov.Replacements {
			replacements[k] = v
		}
		pkgInfo.Replacements = replacements
	}

	if ov.Format != "" {
		pkgInfo.Format = ov.Format
	}

	if ov.Asset != nil {
		pkgInfo.Asset = ov.Asset
	}

	if ov.Files != nil {
		pkgInfo.Files = ov.Files
	}

	if ov.URL != nil {
		pkgInfo.URL = ov.URL
	}
}

type VersionOverride struct {
	Type               string            `json:"type,omitempty" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http"`
	RepoOwner          string            `yaml:"repo_owner" json:"repo_owner,omitempty"`
	RepoName           string            `yaml:"repo_name" json:"repo_name,omitempty"`
	Asset              *string           `json:"asset,omitempty"`
	Path               *string           `json:"path,omitempty"`
	Format             string            `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw"`
	Files              []*File           `json:"files,omitempty"`
	URL                *string           `json:"url,omitempty"`
	Replacements       map[string]string `json:"replacements,omitempty"`
	Overrides          []*Override       `json:"overrides,omitempty"`
	FormatOverrides    []*FormatOverride `yaml:"format_overrides" json:"format_overrides,omitempty"`
	SupportedIf        *string           `yaml:"supported_if" json:"supported_if,omitempty"`
	VersionConstraints string            `yaml:"version_constraint" json:"version_constraint,omitempty"`
	VersionFilter      *string           `yaml:"version_filter" json:"version_filter,omitempty"`
	VersionSource      string            `json:"version_source,omitempty" yaml:"version_source"`
	Rosetta2           *bool             `json:"rosetta2,omitempty"`
}

type Alias struct {
	Name string `json:"name"`
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
	if pkgInfo.Type == PkgInfoTypeGitHubArchive || pkgInfo.Type == PkgInfoTypeGo {
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
	if file.Src == "" {
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

func (pkgInfo *PackageInfo) GetAsset() *string {
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
	case PkgInfoTypeGo:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, "src"), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
		uS, err := pkgInfo.RenderURL(pkg, rt)
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
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGo:
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
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGo:
		return "", nil
	case PkgInfoTypeGitHubContent:
		s, err := pkgInfo.renderTemplateString(*pkgInfo.Path, pkg, rt)
		if err != nil {
			return "", fmt.Errorf("render a package path: %w", err)
		}
		return s, nil
	case PkgInfoTypeGitHubRelease:
		return pkgInfo.renderTemplateString(*pkgInfo.Asset, pkg, rt)
	case PkgInfoTypeHTTP:
		uS, err := pkgInfo.RenderURL(pkg, rt)
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

func (pkgInfo *PackageInfo) renderTemplateString(s string, pkg *Package, rt *runtime.Runtime) (string, error) {
	tpl, err := template.Compile(s)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	return pkgInfo.renderTemplate(tpl, pkg, rt)
}

func (pkgInfo *PackageInfo) renderTemplate(tpl *texttemplate.Template, pkg *Package, rt *runtime.Runtime) (string, error) {
	uS, err := template.ExecuteTemplate(tpl, map[string]interface{}{
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
	return pkgInfo.renderTemplateString(*pkgInfo.URL, pkg, rt)
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
