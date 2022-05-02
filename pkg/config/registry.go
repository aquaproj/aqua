package config

import (
	"errors"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var (
	errInvalidRegistryType = errors.New("registry type is invalid")
	errPathIsRequired      = errors.New("path is required for local registry")
	errRepoOwnerIsRequired = errors.New("repo_owner is required")
	errRepoNameIsRequired  = errors.New("repo_name is required")
	errRefIsRequired       = errors.New("ref is required for github_content registry")
)

type RegistryContentSimple struct {
	PackageInfos []*PackageInfoSimple `yaml:"packages" validate:"dive" json:"packages"`
}

type RegistryContent struct {
	PackageInfos PackageInfos `yaml:"packages" validate:"dive" json:"packages"`
}

type Registry struct {
	Name      string `validate:"required" json:"name,omitempty"`
	Type      string `validate:"required" json:"type,omitempty" jsonschema:"enum=standard,enum=local,enum=github_content"`
	RepoOwner string `yaml:"repo_owner" json:"repo_owner,omitempty"`
	RepoName  string `yaml:"repo_name" json:"repo_name,omitempty"`
	Ref       string `json:"ref,omitempty"`
	Path      string `validate:"required" json:"path,omitempty"`
}

const (
	RegistryTypeGitHubContent = "github_content"
	RegistryTypeLocal         = "local"
	RegistryTypeStandard      = "standard"
)

func (registry *Registry) Validate() error {
	switch registry.Type {
	case RegistryTypeLocal:
		return registry.validateLocal()
	case RegistryTypeGitHubContent:
		return registry.validateGitHubContent()
	default:
		return logerr.WithFields(errInvalidRegistryType, logrus.Fields{ //nolint:wrapcheck
			"registry_type": registry.Type,
		})
	}
}

func (registry *Registry) validateLocal() error {
	if registry.Path == "" {
		return errPathIsRequired
	}
	return nil
}

func (registry *Registry) validateGitHubContent() error {
	if registry.RepoOwner == "" {
		return errRepoOwnerIsRequired
	}
	if registry.RepoName == "" {
		return errRepoNameIsRequired
	}
	if registry.Ref == "" {
		return errRefIsRequired
	}
	return nil
}

func (registry *Registry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Registry
	a := alias(*registry)
	if err := unmarshal(&a); err != nil {
		return err
	}
	if a.Type == RegistryTypeStandard {
		if a.Name == "" {
			a.Name = RegistryTypeStandard
		}
		a.Type = RegistryTypeGitHubContent
		a.RepoOwner = "aquaproj"
		a.RepoName = "aqua-registry"
		a.Path = "registry.yaml"
	}
	*registry = Registry(a)
	return nil
}

func (registry *Registry) GetFilePath(rootDir, cfgFilePath string) string {
	switch registry.Type {
	case RegistryTypeLocal:
		if filepath.IsAbs(registry.Path) {
			return registry.Path
		}
		return filepath.Join(filepath.Dir(cfgFilePath), registry.Path)
	case RegistryTypeGitHubContent:
		return filepath.Join(rootDir, "registries", registry.Type, "github.com", registry.RepoOwner, registry.RepoName, registry.Ref, registry.Path)
	}
	return ""
}
