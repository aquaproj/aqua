package aqua

import (
	"path/filepath"
	"strings"

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

// The number of hex digits of a commit SHA: SHA-1, and SHA-256 for a repository using it.
const (
	sha1HexLen   = 40
	sha256HexLen = 64
)

// CommitFilePath returns where a github_content registry file at a commit the lock file
// recorded is kept.
//
// The file at FilePath is named by the ref, which moves on, and is shared with every
// configuration: another one's lock update rewrites it with whatever the ref points at now.
// A commit doesn't move, so a file kept under it always holds the content of that commit.
// It sits apart from FilePath's tree, where a ref that looks like a commit SHA could
// otherwise collide with it.
//
// The commit comes from the lock file, which nothing validates, so it is checked to be a
// commit SHA before it becomes part of a path.
func (r *Registry) CommitFilePath(rootDir, commit string) (string, error) {
	if r.Type != RegistryTypeGitHubContent {
		return "", slogerr.With(errInvalidRegistryType, "registry_type", r.Type) //nolint:wrapcheck
	}
	commit = strings.ToLower(commit)
	if (len(commit) != sha1HexLen && len(commit) != sha256HexLen) || !isHex(commit) {
		return "", errInvalidCommit
	}
	return filepath.Join(rootDir, "registries", "github_content_commit", "github.com", r.RepoOwner, r.RepoName, commit, r.Path), nil
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
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
