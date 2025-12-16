package aqua

import (
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

// Registry represents a package registry configuration.
// It defines how to access and download package definitions from various sources.
type Registry struct {
	Name      string `json:"name,omitempty"`                                                                     // Registry name identifier
	Type      string `json:"type,omitempty" jsonschema:"enum=standard,enum=local,enum=github_content,enum=http"` // Registry type (standard, local, github_content, http)
	RepoOwner string `yaml:"repo_owner" json:"repo_owner,omitempty"`                                             // GitHub repository owner
	RepoName  string `yaml:"repo_name" json:"repo_name,omitempty"`                                               // GitHub repository name
	Ref       string `json:"ref,omitempty"`                                                                      // Git reference (tag, branch, commit)
	Path      string `json:"path,omitempty"`                                                                     // Path to registry file or directory
	Private   bool   `json:"private,omitempty"`                                                                  // Whether the registry is private
	URL       string `json:"url,omitempty"`                                                                      // URL for http registry
	Version   string `json:"version,omitempty"`                                                                  // Version for http registry
	Format    string `json:"format,omitempty" jsonschema:"enum=raw,enum=tar.gz"`                                 // Format for http registry (raw or tar.gz)
}

// Registry type constants
const (
	// RegistryTypeGitHubContent indicates a registry hosted on GitHub
	RegistryTypeGitHubContent = "github_content"
	// RegistryTypeLocal indicates a registry stored locally on the filesystem
	RegistryTypeLocal = "local"
	// RegistryTypeStandard indicates the default aqua registry
	RegistryTypeStandard = "standard"
	// RegistryTypeHTTP indicates a registry hosted on an HTTP(S) endpoint
	RegistryTypeHTTP = "http"
)

// Validate validates the registry configuration based on its type.
// It ensures all required fields are present and valid for the registry type.
func (r *Registry) Validate() error {
	switch r.Type {
	case RegistryTypeLocal:
		return r.validateLocal()
	case RegistryTypeGitHubContent:
		return r.validateGitHubContent()
	case RegistryTypeHTTP:
		return r.validateHTTP()
	default:
		return logerr.WithFields(errInvalidRegistryType, logrus.Fields{ //nolint:wrapcheck
			"registry_type": r.Type,
		})
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
// The path format depends on the registry type (local vs GitHub content vs HTTP).
func (r *Registry) FilePath(rootDir, cfgFilePath string) (string, error) {
	switch r.Type {
	case RegistryTypeLocal:
		return osfile.Abs(filepath.Dir(cfgFilePath), r.Path), nil
	case RegistryTypeGitHubContent:
		return filepath.Join(rootDir, "registries", r.Type, "github.com", r.RepoOwner, r.RepoName, r.Ref, r.Path), nil
	case RegistryTypeHTTP:
		return r.httpFilePath(rootDir)
	}
	return "", errInvalidRegistryType
}

// RenderURL renders the URL template with the version.
func (r *Registry) RenderURL() (string, error) {
	return template.Execute(r.URL, map[string]any{ //nolint:wrapcheck
		"Version": r.Version,
	})
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

// validateHTTP validates an HTTP registry configuration.
// It ensures all required HTTP fields are present and valid.
func (r *Registry) validateHTTP() error {
	if r.URL == "" {
		return errURLIsRequired
	}
	if r.Version == "" {
		return errVersionIsRequired
	}
	// Prevent path traversal via version parameter
	if strings.Contains(r.Version, "..") {
		return logerr.WithFields(errInvalidVersion, logrus.Fields{ //nolint:wrapcheck
			"version": r.Version,
		})
	}
	if !containsVersionTemplate(r.URL) {
		return errURLMustContainVersion
	}
	if r.Format != "" && r.Format != "raw" && r.Format != "tar.gz" {
		return logerr.WithFields(errInvalidFormat, logrus.Fields{ //nolint:wrapcheck
			"format": r.Format,
		})
	}
	return nil
}

// httpFilePath returns the file system path for an HTTP registry.
// The path includes the registry name and version to ensure uniqueness.
func (r *Registry) httpFilePath(rootDir string) (string, error) {
	// Extract filename from URL or use default based on format
	filename := r.getHTTPFilename()

	return filepath.Join(rootDir, "registries", r.Type, r.Name, r.Version, filename), nil
}

// getHTTPFilename determines the filename for the cached HTTP registry.
func (r *Registry) getHTTPFilename() string {
	if r.Format == "tar.gz" {
		return "registry.tar.gz"
	}
	// Default to registry.yaml for raw format
	return "registry.yaml"
}

// containsVersionTemplate checks if the URL contains {{.Version}}.
func containsVersionTemplate(s string) bool {
	return strings.Contains(s, "{{.Version}}") || strings.Contains(s, "{{ .Version }}")
}
