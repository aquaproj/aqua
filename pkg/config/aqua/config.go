package aqua

import (
	"strings"

	"github.com/invopop/jsonschema"
)

type Package struct {
	Name     string `validate:"required" json:"name,omitempty"`
	Registry string `validate:"required" yaml:",omitempty" json:"registry,omitempty" jsonschema:"description=Registry name,example=foo,default=standard"`
	Version  string `validate:"required" yaml:",omitempty" json:"version,omitempty"`
	Import   string `yaml:",omitempty" json:"import,omitempty"`
}

func (pkg *Package) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Package
	a := alias(*pkg)
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
	*pkg = Package(a)
	if pkg.Registry == "" {
		pkg.Registry = RegistryTypeStandard
	}

	return nil
}

func parseNameWithVersion(name string) (string, string) {
	idx := strings.Index(name, "@")
	if idx == -1 {
		return name, ""
	}
	return name[:idx], name[idx+1:]
}

type Config struct {
	Packages   []*Package `validate:"dive" json:"packages"`
	Registries Registries `validate:"dive" json:"registries"`
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
	PkgInfoTypeGo            = "go"
	PkgInfoTypeGoInstall     = "go_install"
)

func (registries *Registries) UnmarshalYAML(unmarshal func(interface{}) error) error {
	arr := []*Registry{}
	if err := unmarshal(&arr); err != nil {
		return err
	}
	m := make(map[string]*Registry, len(arr))
	for _, registry := range arr {
		m[registry.Name] = registry
	}
	*registries = m
	return nil
}
