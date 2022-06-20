package registry

import (
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/runtime"
)

const (
	PkgInfoTypeGitHubRelease = "github_release"
	PkgInfoTypeGitHubContent = "github_content"
	PkgInfoTypeGitHubArchive = "github_archive"
	PkgInfoTypeHTTP          = "http"
	PkgInfoTypeGo            = "go"
	PkgInfoTypeGoInstall     = "go_install"
)

type PackageInfo struct {
	Name               string             `json:"name,omitempty"`
	Type               string             `validate:"required" json:"type" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install"`
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
	SupportedEnvs      []string           `yaml:"supported_envs" json:"supported_envs,omitempty"`
	VersionFilter      *string            `yaml:"version_filter" json:"version_filter,omitempty"`
	Rosetta2           *bool              `json:"rosetta2,omitempty"`
	Aliases            []*Alias           `json:"aliases,omitempty"`
	VersionSource      string             `json:"version_source,omitempty" yaml:"version_source"`
	CompleteWindowsExt *bool              `json:"complete_windows_ext,omitempty" yaml:"complete_windows_ext"`
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
		SupportedEnvs:      pkgInfo.SupportedEnvs,
		VersionFilter:      pkgInfo.VersionFilter,
		Rosetta2:           pkgInfo.Rosetta2,
		Aliases:            pkgInfo.Aliases,
		VersionSource:      pkgInfo.VersionSource,
		CompleteWindowsExt: pkgInfo.CompleteWindowsExt,
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
	if child.SupportedEnvs != nil {
		pkg.SupportedEnvs = child.SupportedEnvs
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
	if child.CompleteWindowsExt != nil {
		pkg.CompleteWindowsExt = child.CompleteWindowsExt
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

	if ov.CompleteWindowsExt != nil {
		pkgInfo.CompleteWindowsExt = ov.CompleteWindowsExt
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
	SupportedEnvs      []string          `yaml:"supported_envs" json:"supported_envs,omitempty"`
	VersionConstraints string            `yaml:"version_constraint" json:"version_constraint,omitempty"`
	VersionFilter      *string           `yaml:"version_filter" json:"version_filter,omitempty"`
	VersionSource      string            `json:"version_source,omitempty" yaml:"version_source"`
	Rosetta2           *bool             `json:"rosetta2,omitempty"`
	CompleteWindowsExt *bool             `json:"complete_windows_ext,omitempty" yaml:"complete_windows_ext"`
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
	if pkgInfo.Type == PkgInfoTypeGoInstall && pkgInfo.Path != nil {
		return *pkgInfo.Path
	}
	return ""
}

func (pkgInfo *PackageInfo) GetPath() string {
	if pkgInfo.Path != nil {
		return *pkgInfo.Path
	}
	if pkgInfo.Type == PkgInfoTypeGoInstall && pkgInfo.HasRepo() {
		return "github.com/" + pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
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

func (pkgInfo *PackageInfo) Validate() error { //nolint:cyclop
	if pkgInfo.GetName() == "" {
		return errPkgNameIsRequired
	}
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGo:
		if !pkgInfo.HasRepo() {
			return errRepoRequired
		}
		return nil
	case PkgInfoTypeGoInstall:
		if pkgInfo.GetPath() == "" {
			return errGoInstallRequirePath
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
	if pkgInfo.Type == PkgInfoTypeGoInstall {
		if pkgInfo.Asset != nil {
			return []*File{
				{
					Name: *pkgInfo.Asset,
				},
			}
		}
		return []*File{
			{
				Name: filepath.Base(pkgInfo.GetPath()),
			},
		}
	}
	return pkgInfo.Files
}
