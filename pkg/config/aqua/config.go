package aqua

import (
	"strings"

	"github.com/invopop/jsonschema"
)

type Package struct {
	Name          string         `validate:"required" json:"name,omitempty"`
	Registry      string         `validate:"required" yaml:",omitempty" json:"registry,omitempty" jsonschema:"description=Registry name,example=foo,example=local,default=standard"`
	Version       string         `validate:"required" yaml:",omitempty" json:"version,omitempty"`
	Import        string         `yaml:",omitempty" json:"import,omitempty"`
	Tags          []string       `yaml:",omitempty" json:"tags,omitempty"`
	Description   string         `yaml:",omitempty" json:"description,omitempty"`
	Link          string         `yaml:",omitempty" json:"link,omitempty"`
	Update        *Update        `yaml:",omitempty" json:"update,omitempty"`
	FilePath      string         `json:"-" yaml:"-"`
	GoVersionFile string         `json:"go_version_file,omitempty" yaml:"go_version_file,omitempty"`
	Vars          map[string]any `json:"vars,omitempty" yaml:",omitempty"`
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

func (p *Package) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Package
	a := alias(*p)
	if err := unmarshal(&a); err != nil {
		return err
	}
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

	return nil
}

func parseNameWithVersion(name string) (string, string) {
	pkgName, version, _ := strings.Cut(name, "@")
	return pkgName, version
}

type Config struct {
	Packages   []*Package `validate:"dive" json:"packages"`
	Registries Registries `validate:"dive" json:"registries"`
	Checksum   *Checksum  `json:"checksum,omitempty"`
}

type Registries map[string]*Registry

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

func (r *Registries) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
