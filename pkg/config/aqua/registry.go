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

// The number of hex digits of a checksum of each algorithm a lock file may record for a
// registry file.
const (
	sha256HexLen = 64
	sha512HexLen = 128
)

func lockedChecksumLen(algorithm string) int {
	switch algorithm {
	case "sha256":
		return sha256HexLen
	case "sha512":
		return sha512HexLen
	}
	return 0
}

// LockedFilePath returns where a github_content registry file whose checksum a lock file
// recorded is kept.
//
// The file at FilePath is named by the ref, which another configuration's lock update can
// rewrite with whatever the ref points at now. A file kept under its checksum is never
// replaced by different content, so what is read from it is what the lock file recorded.
// It sits apart from FilePath's tree, which a registry path with directories could
// otherwise collide with.
//
// The checksum comes from the lock file, which nothing validates, so it is checked to be
// a hex digest of the algorithm before it becomes part of a path.
func (r *Registry) LockedFilePath(rootDir, algorithm, checksum string) (string, error) {
	if r.Type != RegistryTypeGitHubContent {
		return "", slogerr.With(errInvalidRegistryType, "registry_type", r.Type) //nolint:wrapcheck
	}
	checksum = strings.ToLower(checksum)
	if n := lockedChecksumLen(algorithm); n == 0 || len(checksum) != n || !isHex(checksum) {
		return "", slogerr.With(errInvalidLockedChecksum, "checksum_algorithm", algorithm) //nolint:wrapcheck
	}
	return filepath.Join(rootDir, "registries", "github_content_locked", "github.com", r.RepoOwner, r.RepoName, r.Ref, algorithm+"-"+checksum, r.Path), nil
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
