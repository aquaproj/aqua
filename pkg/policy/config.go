package policy

import (
	"errors"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

const (
	registryTypeStandard = "standard"
)

var (
	errUnknownRegistry     = errors.New("unknown registry")
	errLocalPathIsRequired = errors.New("local registry requires path")
)

type Config struct {
	Path    string
	Allowed bool
	YAML    *ConfigYAML
}

type ConfigYAML struct {
	Registries []*Registry `json:"registries"`
	Packages   []*Package  `json:"packages,omitempty"`
}

type Registry struct {
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"       jsonschema:"enum=standard,enum=local,enum=github_content"`
	RepoOwner string `json:"repo_owner,omitempty" yaml:"repo_owner"`
	RepoName  string `json:"repo_name,omitempty"  yaml:"repo_name"`
	Ref       string `json:"ref,omitempty"`
	Path      string `json:"path,omitempty"`
}

type Package struct {
	Name         string    `json:"name,omitempty"`
	Version      string    `json:"version,omitempty"`
	RegistryName string    `json:"registry,omitempty" yaml:"registry"`
	Registry     *Registry `json:"-"                  yaml:"-"`
}

func (c *Config) Init() error {
	m := make(map[string]*Registry, len(c.YAML.Registries))
	for _, rgst := range c.YAML.Registries {
		if rgst.Type == registryTypeStandard {
			rgst.Type = "github_content"
			rgst.RepoOwner = "aquaproj"
			rgst.RepoName = "aqua-registry"
			if rgst.Path == "" {
				rgst.Path = "registry.yaml"
			}
			if rgst.Name == "" {
				rgst.Name = registryTypeStandard
			}
		}
		if rgst.Type == "local" {
			if rgst.Path == "" {
				return errLocalPathIsRequired
			}
			rgst.Path = osfile.Abs(filepath.Dir(c.Path), rgst.Path)
		}
		m[rgst.Name] = rgst
	}
	for _, pkg := range c.YAML.Packages {
		if pkg.RegistryName == "" {
			pkg.RegistryName = registryTypeStandard
		}
		rgst, ok := m[pkg.RegistryName]
		if !ok {
			return errUnknownRegistry
		}
		pkg.Registry = rgst
	}
	return nil
}
