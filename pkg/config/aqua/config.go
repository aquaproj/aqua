package aqua

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

// Package represents a package definition in aqua.yaml configuration.
// It contains package identification, version constraints, and customization options.
type Package struct {
	Name              string          `json:"name,omitempty"`                                                                                                                                       // Package name
	Registry          string          `yaml:",omitempty"                    json:"registry,omitempty"            jsonschema:"description=Registry name,example=foo,example=local,default=standard"` // Registry containing the package
	Version           string          `yaml:",omitempty"                    json:"version,omitempty"`                                                                                               // Package version
	Import            string          `yaml:",omitempty"                    json:"import,omitempty"`                                                                                                // Import path for configuration inclusion
	Tags              []string        `yaml:",omitempty"                    json:"tags,omitempty"`                                                                                                  // Package tags for filtering
	Description       string          `yaml:",omitempty"                    json:"description,omitempty"`                                                                                           // Package description
	Link              string          `yaml:",omitempty"                    json:"link,omitempty"`                                                                                                  // Package homepage link
	Update            *Update         `yaml:",omitempty"                    json:"update,omitempty"`                                                                                                // Update configuration
	FilePath          string          `yaml:"-"                             json:"-"`                                                                                                               // File path where package is defined
	GoVersionFile     string          `yaml:"go_version_file,omitempty"     json:"go_version_file,omitempty"`                                                                                       // Go version file path
	VersionExpr       string          `yaml:"version_expr,omitempty"        json:"version_expr,omitempty"`                                                                                          // Version expression for dynamic versions
	VersionExprPrefix string          `yaml:"version_expr_prefix,omitempty" json:"version_expr_prefix,omitempty"`                                                                                   // Prefix for version expressions
	Vars              map[string]any  `yaml:",omitempty"                    json:"vars,omitempty"`                                                                                                  // Package-specific variables
	CommandAliases    []*CommandAlias `yaml:"command_aliases,omitempty"     json:"command_aliases,omitempty"`                                                                                       // Command aliases for the package
	Pin               bool            `yaml:"-"                             json:"-"`                                                                                                               // Whether the package version is pinned
}

// CommandAlias defines an alias for a package command.
// It allows creating alternative command names and controls symlink creation.
type CommandAlias struct {
	Command string `json:"command"`                                    // Original command name
	Alias   string `json:"alias"`                                      // Alias name
	NoLink  bool   `yaml:"no_link,omitempty" json:"no_link,omitempty"` // Whether to skip creating symlinks
}

// Update contains configuration for package update behavior.
// It controls whether packages can be updated and what types of updates are allowed.
type Update struct {
	// Enabled controls whether the package is included in update operations.
	// If false, aqua up command ignores the package unless explicitly specified.
	// By default, enabled is true.
	Enabled *bool `yaml:",omitempty" json:"enabled,omitempty"`
	// AllowedVersion is an expression that defines which versions are allowed.
	// Uses expr language (https://github.com/expr-lang/expr).
	// By default, all versions are allowed.
	AllowedVersion string `yaml:"allowed_version,omitempty" json:"allowed_version,omitempty"`
	// Types specifies which update types are allowed (major, minor, patch, other).
	// By default, all types are allowed.
	Types []string `yaml:",omitempty" json:"types,omitempty"`
}

// GetEnabled returns whether updates are enabled for this package.
// It returns true if Update is nil or Enabled is nil (default behavior).
func (u *Update) GetEnabled() bool {
	return u == nil || u.Enabled == nil || *u.Enabled
}

// UnmarshalYAML implements custom YAML unmarshaling for Package.
// It handles package name@version syntax and sets default registry.
func (p *Package) UnmarshalYAML(unmarshal func(any) error) error {
	type alias Package
	a := alias(*p)
	if err := unmarshal(&a); err != nil {
		return err
	}
	pin := a.Version != ""
	name, version := parseNameWithVersion(a.Name)
	if name != "" {
		a.Name = name
	}
	if version != "" {
		a.Version = version
	}
	*p = Package(a)
	if p.Registry == "" {
		p.Registry = RegistryTypeStandard
	}
	p.Pin = pin
	return nil
}

func (p *Package) HasCommandAlias(exeName string) bool {
	for _, a := range p.CommandAliases {
		if a.Alias == exeName {
			return true
		}
	}

	return false
}

// parseNameWithVersion splits a package name containing version (name@version format).
// Returns the package name and version separately.
func parseNameWithVersion(name string) (string, string) {
	pkgName, version, _ := strings.Cut(name, "@")
	return pkgName, version
}

// Config represents the complete aqua.yaml configuration.
// It contains package definitions, registry configurations, and global settings.
type Config struct {
	Packages   []*Package `json:"packages,omitempty"`                               // List of packages to manage
	Registries Registries `json:"registries"`                                       // Registry configurations
	Checksum   *Checksum  `json:"checksum,omitempty"`                               // Checksum validation settings
	ImportDir  string     `yaml:"import_dir,omitempty" json:"import_dir,omitempty"` // Directory for importing configurations
}

// Validate validates the configuration for correctness.
// It checks all registries for proper configuration and returns any validation errors.
func (c *Config) Validate() error {
	for _, r := range c.Registries {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("validate the registry: %w", logerr.WithFields(err, logrus.Fields{
				"registry_name": r.Name,
			}))
		}
	}
	return nil
}

// Registries maps registry names to their configurations.
// It provides custom JSON schema generation for better documentation.
type Registries map[string]*Registry //nolint:recvcheck

// JSONSchema generates a JSON schema for registries configuration.
// It creates an array schema with Registry items for validation purposes.
func (Registries) JSONSchema() *jsonschema.Schema {
	s := jsonschema.Reflect(&Registry{})
	return &jsonschema.Schema{
		Type:  "array",
		Items: s.Definitions["Registry"],
	}
}

// Package type constants for registry package definitions.
const (
	// PkgInfoTypeGitHubRelease indicates packages distributed via GitHub releases
	PkgInfoTypeGitHubRelease = "github_release"
	// PkgInfoTypeGitHubContent indicates packages from GitHub repository content
	PkgInfoTypeGitHubContent = "github_content"
	// PkgInfoTypeGitHubArchive indicates packages using GitHub archive downloads
	PkgInfoTypeGitHubArchive = "github_archive"
	// PkgInfoTypeHTTP indicates packages downloaded from HTTP URLs
	PkgInfoTypeHTTP = "http"
	// PkgInfoTypeGoInstall indicates packages installed via 'go install'
	PkgInfoTypeGoInstall = "go_install"
	// PkgInfoTypeCargo indicates packages installed via Cargo
	PkgInfoTypeCargo = "cargo"
)

// UnmarshalYAML implements custom YAML unmarshaling for Registries.
// It converts from array format to map format for easier lookups.
func (r *Registries) UnmarshalYAML(unmarshal func(any) error) error {
	arr := []*Registry{}
	if err := unmarshal(&arr); err != nil {
		return err
	}
	m := make(map[string]*Registry, len(arr))
	for _, registry := range arr {
		if registry == nil {
			continue
		}
		m[registry.Name] = registry
	}
	*r = m
	return nil
}
