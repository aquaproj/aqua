package aqua

import (
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Registry struct {
	Name      string `validate:"required" json:"name,omitempty"`
	Type      string `validate:"required" json:"type,omitempty" jsonschema:"enum=standard,enum=local,enum=github_content"`
	RepoOwner string `yaml:"repo_owner" json:"repo_owner,omitempty"`
	RepoName  string `yaml:"repo_name" json:"repo_name,omitempty"`
	Ref       string `json:"ref,omitempty"`
	Path      string `validate:"required" json:"path,omitempty"`
	Private   bool   `json:"private,omitempty"`
}

const (
	RegistryTypeGitHubContent = "github_content"
	RegistryTypeLocal         = "local"
	RegistryTypeStandard      = "standard"
)

func (r *Registry) Validate() error {
	switch r.Type {
	case RegistryTypeLocal:
		return r.validateLocal()
	case RegistryTypeGitHubContent:
		return r.validateGitHubContent()
	default:
		return logerr.WithFields(errInvalidRegistryType, logrus.Fields{ //nolint:wrapcheck
			"registry_type": r.Type,
		})
	}
}

func (r *Registry) validateLocal() error {
	if r.Path == "" {
		return errPathIsRequired
	}
	return nil
}

func (r *Registry) validateGitHubContent() error {
	if r.RepoOwner == "" {
		return errRepoOwnerIsRequired
	}
	if r.RepoName == "" {
		return errRepoNameIsRequired
	}
	if r.Ref == "" {
		return errRefIsRequired
	}
	return nil
}

func (r *Registry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Registry
	a := alias(*r)
	if err := unmarshal(&a); err != nil {
		return err
	}
	if a.Type == RegistryTypeStandard {
		a.Type = RegistryTypeGitHubContent
		if a.Name == "" {
			a.Name = RegistryTypeStandard
		}
		if a.RepoOwner == "" {
			a.RepoOwner = "aquaproj"
		}
		if a.RepoName == "" {
			a.RepoName = "aqua-registry"
		}
		if a.Path == "" {
			a.Path = "registry.yaml"
		}
	}
	*r = Registry(a)
	return nil
}

func (r *Registry) GetFilePath(rootDir, cfgFilePath string) (string, error) {
	switch r.Type {
	case RegistryTypeLocal:
		return osfile.Abs(filepath.Dir(cfgFilePath), r.Path), nil
	case RegistryTypeGitHubContent:
		return filepath.Join(rootDir, "registries", r.Type, "github.com", r.RepoOwner, r.RepoName, r.Ref, r.Path), nil
	}
	return "", errInvalidRegistryType
}
