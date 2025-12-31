package aqua

import (
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// Registry represents a package registry configuration.
// It defines how to access and download package definitions from various sources.
type Registry struct {
	Name      string `json:"name,omitempty"`                                                                 // Registry name identifier
	Type      string `json:"type,omitempty"       jsonschema:"enum=standard,enum=local,enum=github_content"` // Registry type (standard, local, github_content)
	RepoOwner string `yaml:"repo_owner" json:"repo_owner,omitempty"`                                         // GitHub repository owner
	RepoName  string `yaml:"repo_name" json:"repo_name,omitempty"`                                           // GitHub repository name
	Ref       string `json:"ref,omitempty"`                                                                  // Git reference (tag, branch, commit)
	Path      string `json:"path,omitempty"`                                                                 // Path to registry file or directory
	Private   bool   `json:"private,omitempty"`                                                              // Whether the registry is private
}

// Registry type constants
const (
	// RegistryTypeGitHubContent indicates a registry hosted on GitHub
	RegistryTypeGitHubContent = "github_content"
	// RegistryTypeLocal indicates a registry stored locally on the filesystem
	RegistryTypeLocal = "local"
	// RegistryTypeStandard indicates the default aqua registry
	RegistryTypeStandard = "standard"
)

// Validate validates the registry configuration based on its type.
// It ensures all required fields are present and valid for the registry type.
func (r *Registry) Validate() error {
	switch r.Type {
	case RegistryTypeLocal:
		return r.validateLocal()
	case RegistryTypeGitHubContent:
		return r.validateGitHubContent()
	default:
		return slogerr.With(errInvalidRegistryType, "registry_type", r.Type) //nolint:wrapcheck
	}
}

// UnmarshalYAML implements custom YAML unmarshaling for Registry.
// It handles the special case of 'standard' registry type with default values.
func (r *Registry) UnmarshalYAML(unmarshal func(any) error) error {
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

// FilePath returns the file system path where the registry file is located.
// The path format depends on the registry type (local vs GitHub content).
func (r *Registry) FilePath(rootDir, cfgFilePath string) (string, error) {
	switch r.Type {
	case RegistryTypeLocal:
		return osfile.Abs(filepath.Dir(cfgFilePath), r.Path), nil
	case RegistryTypeGitHubContent:
		return filepath.Join(rootDir, "registries", r.Type, "github.com", r.RepoOwner, r.RepoName, r.Ref, r.Path), nil
	}
	return "", errInvalidRegistryType
}

// validateLocal validates a local registry configuration.
// It ensures the required path field is present.
func (r *Registry) validateLocal() error {
	if r.Path == "" {
		return errPathIsRequired
	}
	return nil
}

// validateGitHubContent validates a GitHub content registry configuration.
// It ensures all required GitHub fields are present and valid.
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
	if r.Ref == "main" || r.Ref == "master" {
		return errRefCannotBeMainOrMaster
	}
	return nil
}
