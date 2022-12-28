package registry

import (
	"fmt"
	"path"

	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
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
	Name               string             `json:"name,omitempty" yaml:",omitempty"`
	Type               string             `validate:"required" json:"type" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install"`
	RepoOwner          string             `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	RepoName           string             `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	Asset              *string            `json:"asset,omitempty" yaml:",omitempty"`
	Path               *string            `json:"path,omitempty" yaml:",omitempty"`
	Format             string             `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw,example=zip" yaml:",omitempty"`
	Files              []*File            `json:"files,omitempty" yaml:",omitempty"`
	URL                *string            `json:"url,omitempty" yaml:",omitempty"`
	Description        string             `json:"description,omitempty" yaml:",omitempty"`
	Link               string             `json:"link,omitempty" yaml:",omitempty"`
	Replacements       Replacements       `json:"replacements,omitempty" yaml:",omitempty"`
	Overrides          []*Override        `json:"overrides,omitempty" yaml:",omitempty"`
	FormatOverrides    []*FormatOverride  `yaml:"format_overrides,omitempty" json:"format_overrides,omitempty"`
	VersionConstraints string             `yaml:"version_constraint,omitempty" json:"version_constraint,omitempty"`
	VersionOverrides   []*VersionOverride `yaml:"version_overrides,omitempty" json:"version_overrides,omitempty"`
	SupportedIf        *string            `yaml:"supported_if,omitempty" json:"supported_if,omitempty"`
	SupportedEnvs      SupportedEnvs      `yaml:"supported_envs,omitempty" json:"supported_envs,omitempty"`
	VersionFilter      *string            `yaml:"version_filter,omitempty" json:"version_filter,omitempty"`
	Rosetta2           *bool              `yaml:",omitempty" json:"rosetta2,omitempty"`
	Aliases            []*Alias           `yaml:",omitempty" json:"aliases,omitempty"`
	VersionSource      string             `json:"version_source,omitempty" yaml:"version_source,omitempty" jsonschema:"enum=github_tag"`
	CompleteWindowsExt *bool              `json:"complete_windows_ext,omitempty" yaml:"complete_windows_ext,omitempty"`
	WindowsExt         string             `json:"windows_ext,omitempty" yaml:"windows_ext,omitempty"`
	SearchWords        []string           `json:"search_words,omitempty" yaml:"search_words,omitempty"`
	Checksum           *Checksum          `json:"checksum,omitempty"`
	Cosign             *Cosign            `json:"cosign,omitempty"`
	SLSAProvenance     *SLSAProvenance    `json:"slsa_provenance,omitempty" yaml:"slsa_provenance,omitempty"`
}

func (pkgInfo *PackageInfo) Copy() *PackageInfo {
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
		WindowsExt:         pkgInfo.WindowsExt,
		Checksum:           pkgInfo.Checksum,
		Cosign:             pkgInfo.Cosign,
		SLSAProvenance:     pkgInfo.SLSAProvenance,
	}
	return pkg
}

func (pkgInfo *PackageInfo) overrideVersion(child *VersionOverride) *PackageInfo { //nolint:cyclop,funlen
	pkg := pkgInfo.Copy()
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
	if child.WindowsExt != "" {
		pkg.WindowsExt = child.WindowsExt
	}
	if child.Checksum != nil {
		pkg.Checksum = child.Checksum
	}
	if child.Cosign != nil {
		pkg.Cosign = child.Cosign
	}
	if child.SLSAProvenance != nil {
		pkg.SLSAProvenance = child.SLSAProvenance
	}
	return pkg
}

func (pkgInfo *PackageInfo) OverrideByRuntime(rt *runtime.Runtime) { //nolint:cyclop
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
		replacements := make(Replacements, len(pkgInfo.Replacements))
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

	if ov.Checksum != nil {
		pkgInfo.Checksum = ov.Checksum
	}

	if ov.CompleteWindowsExt != nil {
		pkgInfo.CompleteWindowsExt = ov.CompleteWindowsExt
	}
	if ov.WindowsExt != "" {
		pkgInfo.WindowsExt = ov.WindowsExt
	}
	if ov.Type != "" {
		pkgInfo.Type = ov.Type
	}
	if ov.Cosign != nil {
		pkgInfo.Cosign = ov.Cosign
	}
	if ov.SLSAProvenance != nil {
		pkgInfo.SLSAProvenance = ov.SLSAProvenance
	}
}

type VersionOverride struct {
	Type               string            `yaml:",omitempty" json:"type,omitempty" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install"`
	RepoOwner          string            `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	RepoName           string            `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	Asset              *string           `yaml:",omitempty" json:"asset,omitempty"`
	Path               *string           `yaml:",omitempty" json:"path,omitempty"`
	Format             string            `yaml:",omitempty" json:"format,omitempty" jsonschema:"example=tar.gz,example=raw,example=zip"`
	Files              []*File           `yaml:",omitempty" json:"files,omitempty"`
	URL                *string           `yaml:",omitempty" json:"url,omitempty"`
	Replacements       Replacements      `yaml:",omitempty" json:"replacements,omitempty"`
	Overrides          []*Override       `yaml:",omitempty" json:"overrides,omitempty"`
	FormatOverrides    []*FormatOverride `yaml:"format_overrides,omitempty" json:"format_overrides,omitempty"`
	SupportedIf        *string           `yaml:"supported_if,omitempty" json:"supported_if,omitempty"`
	SupportedEnvs      SupportedEnvs     `yaml:"supported_envs,omitempty" json:"supported_envs,omitempty"`
	VersionConstraints string            `yaml:"version_constraint,omitempty" json:"version_constraint,omitempty"`
	VersionFilter      *string           `yaml:"version_filter,omitempty" json:"version_filter,omitempty"`
	VersionSource      string            `json:"version_source,omitempty" yaml:"version_source,omitempty"`
	Rosetta2           *bool             `yaml:",omitempty" json:"rosetta2,omitempty"`
	CompleteWindowsExt *bool             `json:"complete_windows_ext,omitempty" yaml:"complete_windows_ext,omitempty"`
	WindowsExt         string            `json:"windows_ext,omitempty" yaml:"windows_ext,omitempty"`
	Checksum           *Checksum         `json:"checksum,omitempty"`
	Cosign             *Cosign           `json:"cosign,omitempty"`
	SLSAProvenance     *SLSAProvenance   `json:"slsa_provenance,omitempty" yaml:"slsa_provenance,omitempty"`
}

type Alias struct {
	Name string `json:"name"`
}

type Replacements map[string]string

func (Replacements) JSONSchema() *jsonschema.Schema {
	Map := orderedmap.New()
	for _, value := range append(runtime.GOOSList(), runtime.GOARCHList()...) {
		Map.Set(value, &jsonschema.Schema{
			Type: "string",
		})
	}
	return &jsonschema.Schema{
		Type:       "object",
		Properties: Map,
	}
}

type SupportedEnvs []string

func (SupportedEnvs) JSONSchema() *jsonschema.Schema {
	osList := runtime.GOOSList()
	archList := runtime.GOARCHList()
	envs := make([]string, 0, len(osList)*len(archList)+len(osList)+len(archList)+1)
	envs = append(append(append(envs, "all"), osList...), archList...)
	for _, osValue := range runtime.GOOSList() {
		for _, archValue := range runtime.GOARCHList() {
			envs = append(envs, osValue+"/"+archValue)
		}
	}
	s := make([]interface{}, len(envs))
	for i, value := range envs {
		s[i] = value
	}
	return &jsonschema.Schema{
		Type: "array",
		Items: &jsonschema.Schema{
			Type: "string",
			Enum: s,
		},
	}
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

func (pkgInfo *PackageInfo) GetReplacements() Replacements {
	return pkgInfo.Replacements
}

func (pkgInfo *PackageInfo) GetChecksumReplacements() Replacements {
	cr := pkgInfo.Checksum.GetReplacements()
	if cr == nil {
		return pkgInfo.Replacements
	}
	if len(cr) == 0 {
		return cr
	}
	m := Replacements{}
	for k, v := range pkgInfo.Replacements {
		m[k] = v
	}
	for k, v := range cr {
		m[k] = v
	}
	return m
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
				Name: path.Base(pkgInfo.GetPath()),
			},
		}
	}
	return pkgInfo.Files
}

func (pkgInfo *PackageInfo) SLSASourceURI() string {
	sp := pkgInfo.SLSAProvenance
	if sp == nil {
		return ""
	}
	if sp.SourceURI != nil {
		return *sp.SourceURI
	}
	repoOwner := sp.RepoOwner
	repoName := sp.RepoName
	if repoOwner == "" {
		repoOwner = pkgInfo.RepoOwner
	}
	if repoName == "" {
		repoName = pkgInfo.RepoName
	}
	return fmt.Sprintf("github.com/%s/%s", repoOwner, repoName)
}
