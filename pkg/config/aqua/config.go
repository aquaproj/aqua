package aqua

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Package struct {
	Name              string          `json:"name,omitempty"`
	Registry          string          `yaml:",omitempty" json:"registry,omitempty" jsonschema:"description=Registry name,example=foo,example=local,default=standard"`
	Version           string          `yaml:",omitempty" json:"version,omitempty"`
	Import            string          `yaml:",omitempty" json:"import,omitempty"`
	Tags              []string        `yaml:",omitempty" json:"tags,omitempty"`
	Description       string          `yaml:",omitempty" json:"description,omitempty"`
	Link              string          `yaml:",omitempty" json:"link,omitempty"`
	Update            *Update         `yaml:",omitempty" json:"update,omitempty"`
	FilePath          string          `json:"-" yaml:"-"`
	GoVersionFile     string          `json:"go_version_file,omitempty" yaml:"go_version_file,omitempty"`
	VersionExpr       string          `json:"version_expr,omitempty" yaml:"version_expr,omitempty"`
	VersionExprPrefix string          `json:"version_expr_prefix,omitempty" yaml:"version_expr_prefix,omitempty"`
	Vars              map[string]any  `json:"vars,omitempty" yaml:",omitempty"`
	CommandAliases    []*CommandAlias `json:"command_aliases,omitempty" yaml:"command_aliases,omitempty"`
	Pin               bool            `json:"-" yaml:"-"`
}

type CommandAlias struct {
	Command string `json:"command"`
	Alias   string `json:"alias"`
	NoLink  bool   `yaml:"no_link,omitempty" json:"no_link,omitempty"`
}

type Update struct {
	// # If enabled is false, aqua up command ignores the package.
	// If the package name is passed to aqua up command explicitly, enabled is ignored.
	// By default, enabled is true.
	Enabled *bool `yaml:",omitempty" json:"enabled,omitempty"`
	// The condition of allowed version.
	// expr https://github.com/expr-lang/expr is used.
	// By default, all versions are allowed.
	// The version must meet both allowed_version and types.
	AllowedVersion string `yaml:"allowed_version,omitempty" json:"allowed_version,omitempty"`
	// The list of allowed update types
	// By default, all types are allowed.
	// major, minor, patch, other
	Types []string `yaml:",omitempty" json:"types,omitempty"`
}

func (u *Update) GetEnabled() bool {
	return u == nil || u.Enabled == nil || *u.Enabled
}

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

func parseNameWithVersion(name string) (string, string) {
	pkgName, version, _ := strings.Cut(name, "@")
	return pkgName, version
}

type Config struct {
	Packages   []*Package `json:"packages,omitempty"`
	Registries Registries `json:"registries"`
	Checksum   *Checksum  `json:"checksum,omitempty"`
	ImportDir  string     `json:"import_dir,omitempty" yaml:"import_dir,omitempty"`
}

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

type Registries map[string]*Registry //nolint:recvcheck

func (Registries) JSONSchema() *jsonschema.Schema {
	s := jsonschema.Reflect(&Registry{})
	return &jsonschema.Schema{
		Type:  "array",
		Items: s.Definitions["Registry"],
	}
}

const (
	PkgInfoTypeGitHubRelease = "github_release"
	PkgInfoTypeGitHubContent = "github_content"
	PkgInfoTypeGitHubArchive = "github_archive"
	PkgInfoTypeHTTP          = "http"
	PkgInfoTypeGoInstall     = "go_install"
	PkgInfoTypeCargo         = "cargo"
)

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
