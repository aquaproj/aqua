package policy

import (
	"errors"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/util"
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
	Type      string `validate:"required" json:"type,omitempty" jsonschema:"enum=standard,enum=local,enum=github_content"`
	RepoOwner string `yaml:"repo_owner" json:"repo_owner,omitempty"`
	RepoName  string `yaml:"repo_name" json:"repo_name,omitempty"`
	Ref       string `json:"ref,omitempty"`
	Path      string `validate:"required" json:"path,omitempty"`
}

type Package struct {
	Name         string    `json:"name"`
	Version      string    `json:"version,omitempty"`
	RegistryName string    `yaml:"registry" json:"registry,omitempty"`
	Registry     *Registry `yaml:"-" json:"-"`
}

func (cfg *Config) Init() error {
	m := make(map[string]*Registry, len(cfg.YAML.Registries))
	for _, rgst := range cfg.YAML.Registries {
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
			rgst.Path = util.Abs(filepath.Dir(cfg.Path), rgst.Path)
		}
		m[rgst.Name] = rgst
	}
	for _, pkg := range cfg.YAML.Packages {
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
