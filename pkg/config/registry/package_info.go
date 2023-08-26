package registry

import (
	"fmt"
	"path"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
)

const (
	PkgInfoTypeGitHubRelease = "github_release"
	PkgInfoTypeGitHubContent = "github_content"
	PkgInfoTypeGitHubArchive = "github_archive"
	PkgInfoTypeHTTP          = "http"
	PkgInfoTypeGoInstall     = "go_install"
	PkgInfoTypeGoBuild       = "go_build"
	PkgInfoTypeCargo         = "cargo"
)

type PackageInfo struct {
	Name               string             `json:"name,omitempty" yaml:",omitempty"`
	Aliases            []*Alias           `yaml:",omitempty" json:"aliases,omitempty"`
	SearchWords        []string           `json:"search_words,omitempty" yaml:"search_words,omitempty"`
	Type               string             `validate:"required" json:"type" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install,enum=cargo"`
	RepoOwner          string             `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	RepoName           string             `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	Description        string             `json:"description,omitempty" yaml:",omitempty"`
	Link               string             `json:"link,omitempty" yaml:",omitempty"`
	Asset              *string            `json:"asset,omitempty" yaml:",omitempty"`
	Crate              string             `json:"crate,omitempty" yaml:",omitempty"`
	Cargo              *Cargo             `json:"cargo,omitempty"`
	URL                string             `json:"url,omitempty" yaml:",omitempty"`
	Path               string             `json:"path,omitempty" yaml:",omitempty"`
	Format             string             `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw,example=zip,example=dmg" yaml:",omitempty"`
	Overrides          []*Override        `json:"overrides,omitempty" yaml:",omitempty"`
	FormatOverrides    []*FormatOverride  `yaml:"format_overrides,omitempty" json:"format_overrides,omitempty"`
	Files              []*File            `json:"files,omitempty" yaml:",omitempty"`
	Replacements       Replacements       `json:"replacements,omitempty" yaml:",omitempty"`
	SupportedEnvs      SupportedEnvs      `yaml:"supported_envs,omitempty" json:"supported_envs,omitempty"`
	VersionFilter      string             `yaml:"version_filter,omitempty" json:"version_filter,omitempty"`
	VersionPrefix      string             `yaml:"version_prefix,omitempty" json:"version_prefix,omitempty"`
	Rosetta2           *bool              `yaml:",omitempty" json:"rosetta2,omitempty"`
	NoAsset            *bool              `yaml:"no_asset,omitempty" json:"no_asset,omitempty"`
	VersionSource      string             `json:"version_source,omitempty" yaml:"version_source,omitempty" jsonschema:"enum=github_tag"`
	CompleteWindowsExt *bool              `json:"complete_windows_ext,omitempty" yaml:"complete_windows_ext,omitempty"`
	WindowsExt         string             `json:"windows_ext,omitempty" yaml:"windows_ext,omitempty"`
	Checksum           *Checksum          `json:"checksum,omitempty"`
	Cosign             *Cosign            `json:"cosign,omitempty"`
	SLSAProvenance     *SLSAProvenance    `json:"slsa_provenance,omitempty" yaml:"slsa_provenance,omitempty"`
	Private            bool               `json:"private,omitempty"`
	VersionConstraints string             `yaml:"version_constraint,omitempty" json:"version_constraint,omitempty"`
	VersionOverrides   []*VersionOverride `yaml:"version_overrides,omitempty" json:"version_overrides,omitempty"`
	ErrorMessage       string             `json:"-" yaml:"-"`
}

type VersionOverride struct {
	VersionConstraints string          `yaml:"version_constraint,omitempty" json:"version_constraint,omitempty"`
	Type               string          `yaml:",omitempty" json:"type,omitempty" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install"`
	RepoOwner          string          `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	RepoName           string          `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	Asset              *string         `yaml:",omitempty" json:"asset,omitempty"`
	Crate              string          `json:"crate,omitempty" yaml:",omitempty"`
	Cargo              *Cargo          `json:"cargo,omitempty"`
	Path               string          `yaml:",omitempty" json:"path,omitempty"`
	URL                string          `yaml:",omitempty" json:"url,omitempty"`
	Files              []*File         `yaml:",omitempty" json:"files,omitempty"`
	Format             string          `yaml:",omitempty" json:"format,omitempty" jsonschema:"example=tar.gz,example=raw,example=zip"`
	FormatOverrides    FormatOverrides `yaml:"format_overrides,omitempty" json:"format_overrides,omitempty"`
	Overrides          Overrides       `yaml:",omitempty" json:"overrides,omitempty"`
	Replacements       Replacements    `yaml:",omitempty" json:"replacements,omitempty"`
	SupportedEnvs      SupportedEnvs   `yaml:"supported_envs,omitempty" json:"supported_envs,omitempty"`
	VersionFilter      *string         `yaml:"version_filter,omitempty" json:"version_filter,omitempty"`
	VersionPrefix      *string         `yaml:"version_prefix,omitempty" json:"version_prefix,omitempty"`
	VersionSource      string          `json:"version_source,omitempty" yaml:"version_source,omitempty"`
	Rosetta2           *bool           `yaml:",omitempty" json:"rosetta2,omitempty"`
	CompleteWindowsExt *bool           `json:"complete_windows_ext,omitempty" yaml:"complete_windows_ext,omitempty"`
	WindowsExt         string          `json:"windows_ext,omitempty" yaml:"windows_ext,omitempty"`
	Checksum           *Checksum       `json:"checksum,omitempty"`
	Cosign             *Cosign         `json:"cosign,omitempty"`
	SLSAProvenance     *SLSAProvenance `json:"slsa_provenance,omitempty" yaml:"slsa_provenance,omitempty"`
	ErrorMessage       string          `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	NoAsset            *bool           `yaml:"no_asset,omitempty" json:"no_asset,omitempty"`
}

type Override struct {
	GOOS               string          `yaml:",omitempty" json:"goos,omitempty" jsonschema:"enum=aix,enum=android,enum=darwin,enum=dragonfly,enum=freebsd,enum=illumos,enum=ios,enum=linux,enum=netbsd,enum=openbsd,enum=plan9,enum=solaris,enum=windows"`
	GOArch             string          `yaml:",omitempty" json:"goarch,omitempty" jsonschema:"enum=386,enum=amd64,enum=arm,enum=arm64,enum=mips,enum=mips64,enum=mips64le,enum=mipsle,enum=ppc64,enum=ppc64le,enum=riscv64,enum=s390x"`
	Type               string          `json:"type,omitempty" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install"`
	Format             string          `yaml:",omitempty" json:"format,omitempty" jsonschema:"example=tar.gz,example=raw,example=zip"`
	Asset              *string         `yaml:",omitempty" json:"asset,omitempty"`
	Crate              string          `json:"crate,omitempty" yaml:",omitempty"`
	Cargo              *Cargo          `json:"cargo,omitempty"`
	Files              []*File         `yaml:",omitempty" json:"files,omitempty"`
	URL                string          `yaml:",omitempty" json:"url,omitempty"`
	CompleteWindowsExt *bool           `json:"complete_windows_ext,omitempty" yaml:"complete_windows_ext,omitempty"`
	WindowsExt         string          `json:"windows_ext,omitempty" yaml:"windows_ext,omitempty"`
	Replacements       Replacements    `yaml:",omitempty" json:"replacements,omitempty"`
	Checksum           *Checksum       `json:"checksum,omitempty"`
	Cosign             *Cosign         `json:"cosign,omitempty"`
	SLSAProvenance     *SLSAProvenance `json:"slsa_provenance,omitempty" yaml:"slsa_provenance,omitempty"`
}

func (p *PackageInfo) Copy() *PackageInfo {
	pkg := &PackageInfo{
		Name:               p.Name,
		Type:               p.Type,
		RepoOwner:          p.RepoOwner,
		RepoName:           p.RepoName,
		Asset:              p.Asset,
		Crate:              p.Crate,
		Cargo:              p.Cargo,
		Path:               p.Path,
		Format:             p.Format,
		Files:              p.Files,
		URL:                p.URL,
		Description:        p.Description,
		Link:               p.Link,
		Replacements:       p.Replacements,
		Overrides:          p.Overrides,
		FormatOverrides:    p.FormatOverrides,
		VersionConstraints: p.VersionConstraints,
		VersionOverrides:   p.VersionOverrides,
		SupportedEnvs:      p.SupportedEnvs,
		VersionFilter:      p.VersionFilter,
		VersionPrefix:      p.VersionPrefix,
		Rosetta2:           p.Rosetta2,
		Aliases:            p.Aliases,
		VersionSource:      p.VersionSource,
		CompleteWindowsExt: p.CompleteWindowsExt,
		WindowsExt:         p.WindowsExt,
		Checksum:           p.Checksum,
		Cosign:             p.Cosign,
		SLSAProvenance:     p.SLSAProvenance,
		Private:            p.Private,
		ErrorMessage:       p.ErrorMessage,
		NoAsset:            p.NoAsset,
	}
	return pkg
}

func (p *PackageInfo) resetByPkgType(typ string) { //nolint:funlen
	switch typ {
	case PkgInfoTypeGitHubRelease:
		p.URL = ""
		p.Path = ""
		p.Crate = ""
		p.Cargo = nil
	case PkgInfoTypeGitHubContent:
		p.URL = ""
		p.Asset = nil
		p.Crate = ""
		p.Cargo = nil
	case PkgInfoTypeGitHubArchive:
		p.URL = ""
		p.Path = ""
		p.Asset = nil
		p.Crate = ""
		p.Cargo = nil
		p.Format = ""
	case PkgInfoTypeHTTP:
		p.Path = ""
		p.Asset = nil
	case PkgInfoTypeGoInstall:
		p.URL = ""
		p.Asset = nil
		p.Crate = ""
		p.Cargo = nil
		p.WindowsExt = ""
		p.CompleteWindowsExt = nil
		p.Cosign = nil
		p.SLSAProvenance = nil
		p.Format = ""
		p.Rosetta2 = nil
	case PkgInfoTypeGoBuild:
		p.URL = ""
		p.Asset = nil
		p.Crate = ""
		p.Cargo = nil
		p.WindowsExt = ""
		p.CompleteWindowsExt = nil
		p.Cosign = nil
		p.SLSAProvenance = nil
		p.Format = ""
		p.Rosetta2 = nil
	case PkgInfoTypeCargo:
		p.URL = ""
		p.Asset = nil
		p.Path = ""
		p.WindowsExt = ""
		p.CompleteWindowsExt = nil
		p.Cosign = nil
		p.SLSAProvenance = nil
		p.Format = ""
		p.Rosetta2 = nil
	}
}

func (p *PackageInfo) overrideVersion(child *VersionOverride) *PackageInfo { //nolint:cyclop,funlen
	pkg := p.Copy()
	if child.Type != "" {
		pkg.resetByPkgType(child.Type)
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
	if child.Crate != "" {
		pkg.Crate = child.Crate
	}
	if child.Cargo != nil {
		pkg.Cargo = child.Cargo
	}
	if child.Path != "" {
		pkg.Path = child.Path
	}
	if child.Format != "" {
		pkg.Format = child.Format
	}
	if child.Files != nil {
		pkg.Files = child.Files
	}
	if child.URL != "" {
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
	if child.SupportedEnvs != nil {
		pkg.SupportedEnvs = child.SupportedEnvs
	}
	if child.VersionFilter != nil {
		pkg.VersionFilter = *child.VersionFilter
	}
	if child.VersionPrefix != nil {
		pkg.VersionPrefix = *child.VersionPrefix
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
	if child.ErrorMessage != "" {
		pkg.ErrorMessage = child.ErrorMessage
	}
	if child.NoAsset != nil {
		pkg.NoAsset = child.NoAsset
	}
	return pkg
}

func (p *PackageInfo) OverrideByRuntime(rt *runtime.Runtime) { //nolint:cyclop,funlen
	for _, fo := range p.FormatOverrides {
		if fo.GOOS == rt.GOOS {
			p.Format = fo.Format
			break
		}
	}

	ov := p.getOverride(rt)
	if ov == nil {
		return
	}

	if ov.Type != "" {
		p.resetByPkgType(ov.Type)
		p.Type = ov.Type
	}

	if p.Replacements == nil {
		p.Replacements = ov.Replacements
	} else {
		replacements := make(Replacements, len(p.Replacements))
		for k, v := range p.Replacements {
			replacements[k] = v
		}
		for k, v := range ov.Replacements {
			replacements[k] = v
		}
		p.Replacements = replacements
	}

	if ov.Format != "" {
		p.Format = ov.Format
	}

	if ov.Asset != nil {
		p.Asset = ov.Asset
	}

	if ov.Crate != "" {
		p.Crate = ov.Crate
	}

	if ov.Cargo != nil {
		p.Cargo = ov.Cargo
	}

	if ov.Files != nil {
		p.Files = ov.Files
	}

	if ov.URL != "" {
		p.URL = ov.URL
	}

	if ov.Checksum != nil {
		p.Checksum = ov.Checksum
	}

	if ov.CompleteWindowsExt != nil {
		p.CompleteWindowsExt = ov.CompleteWindowsExt
	}
	if ov.WindowsExt != "" {
		p.WindowsExt = ov.WindowsExt
	}
	if ov.Cosign != nil {
		p.Cosign = ov.Cosign
	}
	if ov.SLSAProvenance != nil {
		p.SLSAProvenance = ov.SLSAProvenance
	}
}

type FormatOverrides []*FormatOverride

func (o FormatOverrides) IsZero() bool {
	// Implement yaml.IsZeroer https://pkg.go.dev/gopkg.in/yaml.v3#IsZeroer
	return o == nil
}

type Overrides []*Override

func (o Overrides) IsZero() bool {
	// Implement yaml.IsZeroer https://pkg.go.dev/gopkg.in/yaml.v3#IsZeroer
	return o == nil
}

type Alias struct {
	Name string `json:"name"`
}

type Replacements map[string]string

func (r Replacements) IsZero() bool {
	// Implement yaml.IsZeroer https://pkg.go.dev/gopkg.in/yaml.v3#IsZeroer
	return r == nil
}

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

func (p *PackageInfo) GetRosetta2() bool {
	return p.Rosetta2 != nil && *p.Rosetta2
}

func (p *PackageInfo) HasRepo() bool {
	return p.RepoOwner != "" && p.RepoName != ""
}

func (p *PackageInfo) GetName() string {
	if p.Name != "" {
		return p.Name
	}
	if p.HasRepo() {
		return p.RepoOwner + "/" + p.RepoName
	}
	if p.Type == PkgInfoTypeGoInstall && p.Path != "" {
		return p.Path
	}
	return ""
}

func (p *PackageInfo) GetPath() string {
	if p.Path != "" {
		return p.Path
	}
	if p.Type == PkgInfoTypeGoInstall && p.HasRepo() {
		return "github.com/" + p.RepoOwner + "/" + p.RepoName
	}
	return ""
}

func (p *PackageInfo) GetLink() string {
	if p.Link != "" {
		return p.Link
	}
	if p.HasRepo() {
		return "https://github.com/" + p.RepoOwner + "/" + p.RepoName
	}
	return ""
}

func (p *PackageInfo) GetFormat() string {
	if p.Type == PkgInfoTypeGitHubArchive || p.Type == PkgInfoTypeGoBuild {
		return "tar.gz"
	}
	return p.Format
}

func (p *PackageInfo) GetDescription() string {
	return p.Description
}

func (p *PackageInfo) IsNoAsset() bool {
	return p.NoAsset != nil && *p.NoAsset
}

func (p *PackageInfo) GetType() string {
	return p.Type
}

func (p *PackageInfo) GetReplacements() Replacements {
	return p.Replacements
}

func (p *PackageInfo) GetChecksumReplacements() Replacements {
	cr := p.Checksum.GetReplacements()
	if cr == nil {
		return p.Replacements
	}
	if len(cr) == 0 {
		return cr
	}
	m := Replacements{}
	for k, v := range p.Replacements {
		m[k] = v
	}
	for k, v := range cr {
		m[k] = v
	}
	return m
}

func (p *PackageInfo) GetAsset() *string {
	return p.Asset
}

func (p *PackageInfo) Validate() error { //nolint:cyclop
	if p.GetName() == "" {
		return errPkgNameIsRequired
	}
	switch p.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGoBuild:
		if !p.HasRepo() {
			return errRepoRequired
		}
		return nil
	case PkgInfoTypeGoInstall:
		if p.GetPath() == "" {
			return errGoInstallRequirePath
		}
		return nil
	case PkgInfoTypeCargo:
		if p.Crate == "" {
			return errCargoRequireCrate
		}
		return nil
	case PkgInfoTypeGitHubContent:
		if !p.HasRepo() {
			return errRepoRequired
		}
		if p.Path == "" {
			return errGitHubContentRequirePath
		}
		return nil
	case PkgInfoTypeGitHubRelease:
		if !p.HasRepo() {
			return errRepoRequired
		}
		if p.Asset == nil {
			return errAssetRequired
		}
		return nil
	case PkgInfoTypeHTTP:
		if p.URL == "" {
			return errURLRequired
		}
		return nil
	}
	return errInvalidPackageType
}

func (p *PackageInfo) GetFiles() []*File {
	if len(p.Files) != 0 {
		return p.Files
	}

	if cmdName := p.getDefaultCmdName(); cmdName != "" {
		return []*File{
			{
				Name: cmdName,
			},
		}
	}
	return p.Files
}

func (p *PackageInfo) getDefaultCmdName() string {
	if p.HasRepo() {
		if p.Name == "" {
			return p.RepoName
		}
		if i := strings.LastIndex(p.Name, "/"); i != -1 {
			return p.Name[i+1:]
		}
		return p.Name
	}
	if p.Type == PkgInfoTypeGoInstall {
		if p.Asset != nil {
			return *p.Asset
		}
		return path.Base(p.GetPath())
	}
	return path.Base(p.GetName())
}

func (p *PackageInfo) SLSASourceURI() string {
	sp := p.SLSAProvenance
	if sp == nil {
		return ""
	}
	if sp.SourceURI != nil {
		return *sp.SourceURI
	}
	repoOwner := sp.RepoOwner
	repoName := sp.RepoName
	if repoOwner == "" {
		repoOwner = p.RepoOwner
	}
	if repoName == "" {
		repoName = p.RepoName
	}
	return fmt.Sprintf("github.com/%s/%s", repoOwner, repoName)
}
