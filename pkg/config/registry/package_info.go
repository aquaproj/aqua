// Package registry provides configuration structures and utilities for managing
// package information, including package definitions, verification configurations,
// version handling, platform support, and registry caching.
//
// This package contains the core types used by aqua for:
// - Package metadata and definitions
// - Checksum, signing, and verification configurations
// - Version constraints and overrides
// - Platform support definitions
// - Registry configuration and caching
package registry

import (
	"fmt"
	"maps"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// Package type constants define the supported package installation methods.
const (
	// PkgInfoTypeGitHubRelease installs packages from GitHub release assets.
	PkgInfoTypeGitHubRelease = "github_release"
	// PkgInfoTypeGitHubContent installs packages from specific files in GitHub repositories.
	PkgInfoTypeGitHubContent = "github_content"
	// PkgInfoTypeGitHubArchive installs packages from GitHub repository archives.
	PkgInfoTypeGitHubArchive = "github_archive"
	// PkgInfoTypeHTTP installs packages from arbitrary HTTP URLs.
	PkgInfoTypeHTTP = "http"
	// PkgInfoTypeGoInstall installs Go packages using 'go install'.
	PkgInfoTypeGoInstall = "go_install"
	// PkgInfoTypeGoBuild builds Go packages from source using 'go build'.
	PkgInfoTypeGoBuild = "go_build"
	// PkgInfoTypeCargo installs Rust packages from crates.io using cargo.
	PkgInfoTypeCargo = "cargo"
)

// PackageInfo represents a complete package definition including metadata,
// installation configuration, verification settings, and platform support.
// It contains all information needed to install and verify a package across
// different platforms and versions.
type PackageInfo struct {
	Name                       string                      `yaml:",omitempty"                             json:"name,omitempty"`
	Aliases                    []*Alias                    `yaml:",omitempty"                             json:"aliases,omitempty"`
	SearchWords                []string                    `yaml:"search_words,omitempty"                 json:"search_words,omitempty"`
	Type                       string                      `json:"type"                                   jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install,enum=cargo,enum=go_build"`
	RepoOwner                  string                      `yaml:"repo_owner,omitempty"                   json:"repo_owner,omitempty"`
	RepoName                   string                      `yaml:"repo_name,omitempty"                    json:"repo_name,omitempty"`
	Description                string                      `yaml:",omitempty"                             json:"description,omitempty"`
	Link                       string                      `yaml:",omitempty"                             json:"link,omitempty"`
	Asset                      string                      `yaml:",omitempty"                             json:"asset,omitempty"`
	Crate                      string                      `yaml:",omitempty"                             json:"crate,omitempty"`
	URL                        string                      `yaml:",omitempty"                             json:"url,omitempty"`
	Path                       string                      `yaml:",omitempty"                             json:"path,omitempty"`
	Format                     string                      `yaml:",omitempty"                             json:"format,omitempty"                                                                                                             jsonschema:"example=tar.gz,example=raw,example=zip,example=dmg"`
	VersionFilter              string                      `yaml:"version_filter,omitempty"               json:"version_filter,omitempty"`
	VersionPrefix              string                      `yaml:"version_prefix,omitempty"               json:"version_prefix,omitempty"`
	GoVersionPath              string                      `yaml:"go_version_path,omitempty"              json:"go_version_path,omitempty"`
	Rosetta2                   bool                        `yaml:",omitempty"                             json:"rosetta2,omitempty"`
	WindowsARMEmulation        bool                        `yaml:"windows_arm_emulation,omitempty"        json:"windows_arm_emulation,omitempty"`
	NoAsset                    bool                        `yaml:"no_asset,omitempty"                     json:"no_asset,omitempty"`
	VersionSource              string                      `yaml:"version_source,omitempty"               json:"version_source,omitempty"                                                                                                     jsonschema:"enum=github_tag"`
	CompleteWindowsExt         *bool                       `yaml:"complete_windows_ext,omitempty"         json:"complete_windows_ext,omitempty"`
	WindowsExt                 string                      `yaml:"windows_ext,omitempty"                  json:"windows_ext,omitempty"`
	Private                    bool                        `json:"private,omitempty"`
	ErrorMessage               string                      `yaml:"-"                                      json:"-"`
	AppendExt                  *bool                       `yaml:"append_ext,omitempty"                   json:"append_ext,omitempty"`
	Cargo                      *Cargo                      `json:"cargo,omitempty"`
	Build                      *Build                      `yaml:",omitempty"                             json:"build,omitempty"`
	Overrides                  []*Override                 `yaml:",omitempty"                             json:"overrides,omitempty"`
	FormatOverrides            []*FormatOverride           `yaml:"format_overrides,omitempty"             json:"format_overrides,omitempty"`
	Files                      []*File                     `yaml:",omitempty"                             json:"files,omitempty"`
	Replacements               Replacements                `yaml:",omitempty"                             json:"replacements,omitempty"`
	SupportedEnvs              SupportedEnvs               `yaml:"supported_envs,omitempty"               json:"supported_envs,omitempty"`
	Checksum                   *Checksum                   `json:"checksum,omitempty"`
	Cosign                     *Cosign                     `json:"cosign,omitempty"`
	SLSAProvenance             *SLSAProvenance             `yaml:"slsa_provenance,omitempty"              json:"slsa_provenance,omitempty"`
	Minisign                   *Minisign                   `yaml:",omitempty"                             json:"minisign,omitempty"`
	GitHubArtifactAttestations *GitHubArtifactAttestations `yaml:"github_artifact_attestations,omitempty" json:"github_artifact_attestations,omitempty"`
	GitHubImmutableRelease     bool                        `yaml:"github_immutable_release,omitempty"     json:"github_immutable_release,omitempty"`
	Vars                       []*Var                      `yaml:",omitempty"                             json:"vars,omitempty"`
	VersionConstraints         string                      `yaml:"version_constraint,omitempty"           json:"version_constraint,omitempty"`
	VersionOverrides           []*VersionOverride          `yaml:"version_overrides,omitempty"            json:"version_overrides,omitempty"`
}

// Var represents a template variable that can be used in package configurations
// to customize installation behavior based on runtime values.
type Var struct {
	// Name is the variable name used in templates.
	Name string `json:"name"`
	// Required indicates whether this variable must be provided.
	Required bool `json:"required,omitempty"`
	// Default is the default value used when the variable is not provided.
	Default any `json:"default,omitempty"`
}

// Build defines configuration for building packages from source code.
// This is used as a fallback when pre-built binaries are not available
// for the target platform.
type Build struct {
	// Enabled controls whether building from source is allowed.
	Enabled *bool `yaml:",omitempty" json:"enabled,omitempty"`
	// Type specifies the build method (go_install or go_build).
	Type string `yaml:",omitempty" json:"type,omitempty" jsonschema:"enum=go_install,enum=go_build"`
	// Path is the import path or directory for Go packages.
	Path string `yaml:",omitempty" json:"path,omitempty"`
	// Files specifies which files to install after building.
	Files []*File `yaml:",omitempty" json:"files,omitempty"`
	// ExcludedEnvs lists environments where building should be skipped.
	ExcludedEnvs SupportedEnvs `yaml:"excluded_envs,omitempty" json:"excluded_envs,omitempty"`
}

// CheckEnabled returns true if building from source is enabled for this package.
// If Enabled is nil, it defaults to true.
func (b *Build) CheckEnabled() bool {
	if b == nil {
		return false
	}
	if b.Enabled == nil {
		return true
	}
	return *b.Enabled
}

// GetAppendExt returns whether file extensions should be appended to installed binaries.
// If AppendExt is nil, it defaults to true.
func (p *PackageInfo) GetAppendExt() bool {
	if p.AppendExt == nil {
		return true
	}
	return *p.AppendExt
}

// VersionOverride allows different package configurations for specific version ranges.
// This enables packages to change their installation method, repository, or other
// settings based on the version being installed.
type VersionOverride struct {
	VersionConstraints         string                      `yaml:"version_constraint,omitempty"           json:"version_constraint,omitempty"`
	Type                       string                      `yaml:",omitempty"                             json:"type,omitempty"                         jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install,enum=cargo,enum=go_build"`
	RepoOwner                  string                      `yaml:"repo_owner,omitempty"                   json:"repo_owner,omitempty"`
	RepoName                   string                      `yaml:"repo_name,omitempty"                    json:"repo_name,omitempty"`
	Asset                      string                      `yaml:",omitempty"                             json:"asset,omitempty"`
	Crate                      string                      `yaml:",omitempty"                             json:"crate,omitempty"`
	Path                       string                      `yaml:",omitempty"                             json:"path,omitempty"`
	URL                        string                      `yaml:",omitempty"                             json:"url,omitempty"`
	Format                     string                      `yaml:",omitempty"                             json:"format,omitempty"                       jsonschema:"example=tar.gz,example=raw,example=zip"`
	GoVersionPath              *string                     `yaml:"go_version_path,omitempty"              json:"go_version_path,omitempty"`
	VersionFilter              *string                     `yaml:"version_filter,omitempty"               json:"version_filter,omitempty"`
	VersionPrefix              *string                     `yaml:"version_prefix,omitempty"               json:"version_prefix,omitempty"`
	VersionSource              string                      `yaml:"version_source,omitempty"               json:"version_source,omitempty"`
	WindowsExt                 string                      `yaml:"windows_ext,omitempty"                  json:"windows_ext,omitempty"`
	ErrorMessage               *string                     `yaml:"error_message,omitempty"                json:"error_message,omitempty"`
	Rosetta2                   *bool                       `yaml:",omitempty"                             json:"rosetta2,omitempty"`
	WindowsARMEmulation        *bool                       `yaml:"windows_arm_emulation,omitempty"        json:"windows_arm_emulation,omitempty"`
	CompleteWindowsExt         *bool                       `yaml:"complete_windows_ext,omitempty"         json:"complete_windows_ext,omitempty"`
	NoAsset                    *bool                       `yaml:"no_asset,omitempty"                     json:"no_asset,omitempty"`
	AppendExt                  *bool                       `yaml:"append_ext,omitempty"                   json:"append_ext,omitempty"`
	Cargo                      *Cargo                      `json:"cargo,omitempty"`
	Files                      []*File                     `yaml:",omitempty"                             json:"files,omitempty"`
	FormatOverrides            FormatOverrides             `yaml:"format_overrides,omitempty"             json:"format_overrides,omitempty"`
	Replacements               Replacements                `yaml:",omitempty"                             json:"replacements,omitempty"`
	Checksum                   *Checksum                   `json:"checksum,omitempty"`
	Cosign                     *Cosign                     `json:"cosign,omitempty"`
	SLSAProvenance             *SLSAProvenance             `yaml:"slsa_provenance,omitempty"              json:"slsa_provenance,omitempty"`
	Minisign                   *Minisign                   `yaml:",omitempty"                             json:"minisign,omitempty"`
	GitHubArtifactAttestations *GitHubArtifactAttestations `yaml:"github_artifact_attestations,omitempty" json:"github_artifact_attestations,omitempty"`
	GitHubImmutableRelease     *bool                       `yaml:"github_immutable_release,omitempty"     json:"github_immutable_release,omitempty"`
	Build                      *Build                      `yaml:",omitempty"                             json:"build,omitempty"`
	Vars                       []*Var                      `yaml:",omitempty"                             json:"vars,omitempty"`
	Overrides                  Overrides                   `yaml:",omitempty"                             json:"overrides,omitempty"`
	SupportedEnvs              SupportedEnvs               `yaml:"supported_envs,omitempty"               json:"supported_envs,omitempty"`
}

// Override provides platform-specific package configuration that overrides
// the default settings when the specified OS/architecture conditions are met.
type Override struct {
	GOOS                       string                      `yaml:",omitempty"                             json:"goos,omitempty"                                                                                                               jsonschema:"enum=darwin,enum=linux,enum=windows"`
	GOArch                     string                      `yaml:",omitempty"                             json:"goarch,omitempty"                                                                                                             jsonschema:"enum=amd64,enum=arm64"`
	Type                       string                      `json:"type,omitempty"                         jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install,enum=cargo,enum=go_build"`
	Format                     string                      `yaml:",omitempty"                             json:"format,omitempty"                                                                                                             jsonschema:"example=tar.gz,example=raw,example=zip"`
	Asset                      string                      `yaml:",omitempty"                             json:"asset,omitempty"`
	Crate                      string                      `yaml:",omitempty"                             json:"crate,omitempty"`
	URL                        string                      `yaml:",omitempty"                             json:"url,omitempty"`
	Path                       string                      `yaml:",omitempty"                             json:"path,omitempty"`
	GoVersionPath              *string                     `yaml:"go_version_path,omitempty"              json:"go_version_path,omitempty"`
	CompleteWindowsExt         *bool                       `yaml:"complete_windows_ext,omitempty"         json:"complete_windows_ext,omitempty"`
	WindowsExt                 string                      `yaml:"windows_ext,omitempty"                  json:"windows_ext,omitempty"`
	AppendExt                  *bool                       `yaml:"append_ext,omitempty"                   json:"append_ext,omitempty"`
	Cargo                      *Cargo                      `json:"cargo,omitempty"`
	Files                      []*File                     `yaml:",omitempty"                             json:"files,omitempty"`
	Replacements               Replacements                `yaml:",omitempty"                             json:"replacements,omitempty"`
	Checksum                   *Checksum                   `json:"checksum,omitempty"`
	Cosign                     *Cosign                     `json:"cosign,omitempty"`
	SLSAProvenance             *SLSAProvenance             `yaml:"slsa_provenance,omitempty"              json:"slsa_provenance,omitempty"`
	Minisign                   *Minisign                   `yaml:",omitempty"                             json:"minisign,omitempty"`
	GitHubArtifactAttestations *GitHubArtifactAttestations `yaml:"github_artifact_attestations,omitempty" json:"github_artifact_attestations,omitempty"`
	Vars                       []*Var                      `yaml:",omitempty"                             json:"vars,omitempty"`
	Envs                       SupportedEnvs               `yaml:",omitempty"                             json:"envs,omitempty"`
}

// Copy creates a deep copy of the PackageInfo struct.
// This is used when creating modified versions of a package
// for different platforms or versions.
func (p *PackageInfo) Copy() *PackageInfo {
	pkg := &PackageInfo{
		Name:                       p.Name,
		Type:                       p.Type,
		RepoOwner:                  p.RepoOwner,
		RepoName:                   p.RepoName,
		Asset:                      p.Asset,
		Crate:                      p.Crate,
		Cargo:                      p.Cargo,
		Path:                       p.Path,
		Format:                     p.Format,
		Files:                      p.Files,
		URL:                        p.URL,
		Description:                p.Description,
		Link:                       p.Link,
		Replacements:               p.Replacements,
		Overrides:                  p.Overrides,
		FormatOverrides:            p.FormatOverrides,
		VersionConstraints:         p.VersionConstraints,
		VersionOverrides:           p.VersionOverrides,
		SupportedEnvs:              p.SupportedEnvs,
		VersionFilter:              p.VersionFilter,
		VersionPrefix:              p.VersionPrefix,
		GoVersionPath:              p.GoVersionPath,
		Rosetta2:                   p.Rosetta2,
		WindowsARMEmulation:        p.WindowsARMEmulation,
		Aliases:                    p.Aliases,
		VersionSource:              p.VersionSource,
		CompleteWindowsExt:         p.CompleteWindowsExt,
		WindowsExt:                 p.WindowsExt,
		Checksum:                   p.Checksum,
		Cosign:                     p.Cosign,
		SLSAProvenance:             p.SLSAProvenance,
		Minisign:                   p.Minisign,
		GitHubArtifactAttestations: p.GitHubArtifactAttestations,
		GitHubImmutableRelease:     p.GitHubImmutableRelease,
		Private:                    p.Private,
		ErrorMessage:               p.ErrorMessage,
		NoAsset:                    p.NoAsset,
		AppendExt:                  p.AppendExt,
		Build:                      p.Build,
		Vars:                       p.Vars,
	}
	return pkg
}

// OverrideByRuntime applies platform-specific overrides based on the runtime environment.
// It modifies the PackageInfo in-place to use platform-specific settings when available.
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

	if ov.Asset != "" {
		p.Asset = ov.Asset
	}

	if ov.Crate != "" {
		p.Crate = ov.Crate
	}

	if ov.Cargo != nil {
		p.Cargo = ov.Cargo
	}

	if ov.URL != "" {
		p.URL = ov.URL
	}

	if ov.Path != "" {
		p.Path = ov.Path
	}

	if ov.Format != "" {
		p.Format = ov.Format
	}

	if ov.Files != nil {
		p.Files = ov.Files
	}

	if p.Replacements == nil {
		p.Replacements = ov.Replacements
	} else {
		replacements := make(Replacements, len(p.Replacements))
		maps.Copy(replacements, p.Replacements)
		maps.Copy(replacements, ov.Replacements)
		p.Replacements = replacements
	}

	if ov.CompleteWindowsExt != nil {
		p.CompleteWindowsExt = ov.CompleteWindowsExt
	}

	if ov.WindowsExt != "" {
		p.WindowsExt = ov.WindowsExt
	}

	if ov.Checksum != nil {
		p.Checksum = ov.Checksum
	}

	if ov.Cosign != nil {
		p.Cosign = ov.Cosign
	}

	if ov.SLSAProvenance != nil {
		p.SLSAProvenance = ov.SLSAProvenance
	}

	if ov.Minisign != nil {
		p.Minisign = ov.Minisign
	}

	if ov.GitHubArtifactAttestations != nil {
		p.GitHubArtifactAttestations = ov.GitHubArtifactAttestations
	}

	if ov.AppendExt != nil {
		p.AppendExt = ov.AppendExt
	}

	if ov.Vars != nil {
		p.Vars = ov.Vars
	}

	if ov.GoVersionPath != nil {
		p.GoVersionPath = *ov.GoVersionPath
	}
}

// OverrideByBuild applies build-specific configuration to the package.
// This modifies the package to use build settings when building from source.
func (p *PackageInfo) OverrideByBuild() {
	if p.Type != p.Build.Type {
		p.resetByPkgType(p.Build.Type)
		p.Type = p.Build.Type
	}
	if p.Build.Path != "" {
		p.Path = p.Build.Path
	}
	if p.Build.Files != nil {
		p.Files = p.Build.Files
	}
}

// FormatOverrides is a slice of platform-specific format overrides.
type FormatOverrides []*FormatOverride

// IsZero implements yaml.IsZeroer interface.
// It returns true if the FormatOverrides slice is nil.
func (o FormatOverrides) IsZero() bool {
	// Implement yaml.IsZeroer https://pkg.go.dev/gopkg.in/yaml.v3#IsZeroer
	return o == nil
}

// Overrides is a slice of platform-specific configuration overrides.
type Overrides []*Override

// IsZero implements yaml.IsZeroer interface.
// It returns true if the Overrides slice is nil.
func (o Overrides) IsZero() bool {
	// Implement yaml.IsZeroer https://pkg.go.dev/gopkg.in/yaml.v3#IsZeroer
	return o == nil
}

// Alias represents an alternative name for a package.
// This allows packages to be referenced by multiple names.
type Alias struct {
	// Name is the alternative name for the package.
	Name string `json:"name"`
}

// Replacements is a map of template replacements for platform-specific values.
// Keys are typically platform identifiers (GOOS/GOARCH) and values are the replacements.
type Replacements map[string]string

// IsZero implements yaml.IsZeroer interface.
// It returns true if the Replacements map is nil.
func (r Replacements) IsZero() bool {
	// Implement yaml.IsZeroer https://pkg.go.dev/gopkg.in/yaml.v3#IsZeroer
	return r == nil
}

// JSONSchema generates a JSON schema for Replacements.
// It creates a schema with properties for each supported GOOS and GOARCH value.
func (Replacements) JSONSchema() *jsonschema.Schema {
	Map := orderedmap.New[string, *jsonschema.Schema]()
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

// SupportedEnvs represents a list of supported runtime environments.
// Each entry can be a GOOS, GOARCH, GOOS/GOARCH combination, or "all".
type SupportedEnvs []string

// JSONSchema generates a JSON schema for SupportedEnvs.
// It creates an enum schema with all valid GOOS, GOARCH, and GOOS/GOARCH combinations.
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
	s := make([]any, len(envs))
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

// HasRepo returns true if the package has both RepoOwner and RepoName set.
// This indicates the package is hosted on a Git repository (typically GitHub).
func (p *PackageInfo) HasRepo() bool {
	return p.RepoOwner != "" && p.RepoName != ""
}

// GetName returns the effective name of the package.
// It uses the Name field if set, otherwise derives it from repository or path information.
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

// GetPath returns the effective import path or directory path for the package.
// For Go packages, this is used with 'go install' or 'go build'.
func (p *PackageInfo) GetPath() string {
	if p.Path != "" {
		return p.Path
	}
	if p.Type == PkgInfoTypeGoInstall && p.HasRepo() {
		return "github.com/" + p.RepoOwner + "/" + p.RepoName
	}
	return ""
}

// GetLink returns the primary URL link for the package.
// It uses the Link field if set, otherwise derives it from repository information.
func (p *PackageInfo) GetLink() string {
	if p.Link != "" {
		return p.Link
	}
	if p.HasRepo() {
		return "https://github.com/" + p.RepoOwner + "/" + p.RepoName
	}
	return ""
}

// GetFormat returns the archive format for the package.
// Some package types have default formats that are returned if Format is not explicitly set.
func (p *PackageInfo) GetFormat() string {
	if p.Type == PkgInfoTypeGitHubArchive || p.Type == PkgInfoTypeGoBuild {
		return "tar.gz"
	}
	return p.Format
}

// GetChecksumReplacements returns the effective replacements for checksum validation.
// It merges package-level replacements with checksum-specific replacements.
func (p *PackageInfo) GetChecksumReplacements() Replacements {
	cr := p.Checksum.GetReplacements()
	if cr == nil {
		return p.Replacements
	}
	if len(cr) == 0 {
		return cr
	}
	m := Replacements{}
	maps.Copy(m, p.Replacements)
	maps.Copy(m, cr)
	return m
}

// Validate checks if the PackageInfo has all required fields for its type.
// It returns an error if any required fields are missing or invalid.
func (p *PackageInfo) Validate() error { //nolint:cyclop
	if p.GetName() == "" {
		return errPkgNameIsRequired
	}
	if p.NoAsset || p.ErrorMessage != "" {
		return nil
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
		if p.Asset == "" {
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

// GetFiles returns the list of files to be installed.
// If no files are specified, it returns a default file based on the package name.
func (p *PackageInfo) GetFiles() []*File {
	if len(p.Files) != 0 {
		return p.Files
	}

	if cmdName := p.defaultCmdName(); cmdName != "" {
		return []*File{
			{
				Name: cmdName,
			},
		}
	}
	return p.Files
}

func doFilesContain(files []*File, exeName string, isEmpty *bool) bool {
	if len(files) == 0 {
		*isEmpty = true
	}

	for _, f := range files {
		if f.Name == exeName {
			return true
		}
	}

	return false
}

// MaybeHasCommand returns true if the given exe name can be in this package.
// This includes file lists that may only be used under specific versions or
// host platforms.
func (p *PackageInfo) MaybeHasCommand(exeName string) bool { //nolint:cyclop
	anyListEmpty := false

	if doFilesContain(p.Files, exeName, &anyListEmpty) {
		return true
	}

	if p.Build != nil && doFilesContain(p.Build.Files, exeName, &anyListEmpty) {
		return true
	}

	for _, v := range p.VersionOverrides {
		if doFilesContain(v.Files, exeName, &anyListEmpty) {
			return true
		}

		if v.Build != nil && doFilesContain(v.Build.Files, exeName, &anyListEmpty) {
			return true
		}

		for _, o := range v.Overrides {
			if doFilesContain(o.Files, exeName, &anyListEmpty) {
				return true
			}
		}
	}

	for _, o := range p.Overrides {
		if doFilesContain(o.Files, exeName, &anyListEmpty) {
			return true
		}
	}

	// If any of the file lists that could be used are empty, then the default
	// command name would be used, so check that as well.
	if anyListEmpty && p.defaultCmdName() == exeName {
		return true
	}

	return false
}

var placeHolderTemplate = regexp.MustCompile(`{{.*?}}`)

// PkgPaths returns all possible package installation paths for this package.
// This includes paths for all version overrides and is used for package management.
func (p *PackageInfo) PkgPaths() map[string]struct{} {
	m := map[string]struct{}{}
	for _, a := range p.pkgPaths() {
		m[a] = struct{}{}
	}
	for _, vo := range p.VersionOverrides {
		pkg := p.overrideVersion(vo)
		for _, a := range pkg.pkgPaths() {
			m[a] = struct{}{}
		}
	}
	return m
}

// SLSASourceURI returns the source URI for SLSA provenance verification.
// It uses the configured SourceURI or derives it from repository information.
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

// pkgPaths returns the package installation paths for this specific package configuration.
// This is used internally by PkgPaths to compute all possible paths.
func (p *PackageInfo) pkgPaths() []string { //nolint:cyclop
	if p.NoAsset || p.ErrorMessage != "" {
		return nil
	}
	switch p.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGoBuild, PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		if p.RepoOwner == "" || p.RepoName == "" {
			return nil
		}
		return []string{filepath.Join(p.Type, "github.com", p.RepoOwner, p.RepoName)}
	case PkgInfoTypeCargo:
		if p.Crate == "" {
			return nil
		}
		return []string{filepath.Join(p.Type, "crates.io", p.Crate)}
	case PkgInfoTypeGoInstall:
		a := p.GetPath()
		if a == "" {
			return nil
		}
		return []string{filepath.Join(p.Type, filepath.FromSlash(placeHolderTemplate.ReplaceAllLiteralString(a, "*")))}
	case PkgInfoTypeHTTP:
		if p.URL == "" {
			return nil
		}
		u, err := url.Parse(placeHolderTemplate.ReplaceAllLiteralString(p.URL, "*"))
		if err != nil {
			return nil
		}
		return []string{filepath.Join(p.Type, u.Host, filepath.FromSlash(u.Path))}
	}
	return nil
}

// defaultCmdName returns the default command name for the package.
// This is derived from the package name, repository name, or path.
func (p *PackageInfo) defaultCmdName() string {
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
		return path.Base(p.GetPath())
	}
	return path.Base(p.GetName())
}

// overrideVersion applies a version override to create a new PackageInfo.
// This creates a copy with version-specific configuration applied.
func (p *PackageInfo) overrideVersion(child *VersionOverride) *PackageInfo { //nolint:cyclop,funlen,gocyclo,gocognit
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
	if child.Asset != "" {
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
	if child.GoVersionPath != nil {
		pkg.GoVersionPath = *child.GoVersionPath
	}
	if child.Rosetta2 != nil {
		pkg.Rosetta2 = *child.Rosetta2
	}
	if child.WindowsARMEmulation != nil {
		pkg.WindowsARMEmulation = *child.WindowsARMEmulation
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
	if child.Minisign != nil {
		pkg.Minisign = child.Minisign
	}
	if child.GitHubArtifactAttestations != nil {
		pkg.GitHubArtifactAttestations = child.GitHubArtifactAttestations
	}
	if child.GitHubImmutableRelease != nil {
		pkg.GitHubImmutableRelease = *child.GitHubImmutableRelease
	}
	if child.ErrorMessage != nil {
		pkg.ErrorMessage = *child.ErrorMessage
	}
	if child.NoAsset != nil {
		pkg.NoAsset = *child.NoAsset
	}
	if child.AppendExt != nil {
		pkg.AppendExt = child.AppendExt
	}
	if child.Build != nil {
		pkg.Build = child.Build
	}
	if child.Vars != nil {
		pkg.Vars = child.Vars
	}
	return pkg
}

// resetByPkgType resets package fields that are not applicable to the specified type.
// This cleans up conflicting configuration when changing package types.
func (p *PackageInfo) resetByPkgType(typ string) { //nolint:funlen
	switch typ {
	case PkgInfoTypeGitHubRelease:
		p.URL = ""
		p.Path = ""
		p.Crate = ""
		p.GoVersionPath = ""
		p.Cargo = nil
	case PkgInfoTypeGitHubContent:
		p.URL = ""
		p.Asset = ""
		p.Crate = ""
		p.GoVersionPath = ""
		p.Cargo = nil
	case PkgInfoTypeGitHubArchive:
		p.URL = ""
		p.Path = ""
		p.Asset = ""
		p.Crate = ""
		p.GoVersionPath = ""
		p.Cargo = nil
		p.Format = ""
	case PkgInfoTypeHTTP:
		p.Path = ""
		p.Asset = ""
		p.GoVersionPath = ""
	case PkgInfoTypeGoInstall:
		p.URL = ""
		p.Asset = ""
		p.Crate = ""
		p.Cargo = nil
		p.WindowsExt = ""
		p.CompleteWindowsExt = nil
		p.Cosign = nil
		p.SLSAProvenance = nil
		p.Minisign = nil
		p.GitHubArtifactAttestations = nil
		p.GitHubImmutableRelease = false
		p.Format = ""
		p.Rosetta2 = false
		p.WindowsARMEmulation = false
		p.AppendExt = nil
	case PkgInfoTypeGoBuild:
		p.URL = ""
		p.Asset = ""
		p.Crate = ""
		p.Cargo = nil
		p.WindowsExt = ""
		p.CompleteWindowsExt = nil
		p.Cosign = nil
		p.SLSAProvenance = nil
		p.Minisign = nil
		p.GitHubArtifactAttestations = nil
		p.GitHubImmutableRelease = false
		p.Format = ""
		p.Rosetta2 = false
		p.WindowsARMEmulation = false
		p.AppendExt = nil
	case PkgInfoTypeCargo:
		p.URL = ""
		p.Asset = ""
		p.Path = ""
		p.WindowsExt = ""
		p.CompleteWindowsExt = nil
		p.Cosign = nil
		p.SLSAProvenance = nil
		p.Minisign = nil
		p.GitHubArtifactAttestations = nil
		p.GitHubImmutableRelease = false
		p.Format = ""
		p.Rosetta2 = false
		p.WindowsARMEmulation = false
		p.AppendExt = nil
		p.GoVersionPath = ""
	}
}
