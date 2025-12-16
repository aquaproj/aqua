package aqua

import "errors"

// Error variables for registry validation
var (
	// errInvalidRegistryType is returned when an unsupported registry type is specified
	errInvalidRegistryType = errors.New("registry type is invalid")
	// errPathIsRequired is returned when a local registry doesn't specify a path
	errPathIsRequired = errors.New("path is required for local registry")
	// errRepoOwnerIsRequired is returned when a GitHub registry lacks repo_owner
	errRepoOwnerIsRequired = errors.New("repo_owner is required")
	// errRepoNameIsRequired is returned when a GitHub registry lacks repo_name
	errRepoNameIsRequired = errors.New("repo_name is required")
	// errRefIsRequired is returned when github_content registry doesn't specify ref
	errRefIsRequired = errors.New("ref is required for github_content registry")
	// errRefCannotBeMainOrMaster is returned when github_content registry uses unstable refs
	errRefCannotBeMainOrMaster = errors.New("ref cannot be 'main' or 'master' for github_content registry")
	// errURLIsRequired is returned when an HTTP registry doesn't specify a URL
	errURLIsRequired = errors.New("url is required for http registry")
	// errVersionIsRequired is returned when an HTTP registry doesn't specify a version
	errVersionIsRequired = errors.New("version is required for http registry")
	// errURLMustContainVersion is returned when an HTTP registry URL doesn't contain {{.Version}}
	errURLMustContainVersion = errors.New("url must contain {{.Version}} template for http registry")
	// errInvalidFormat is returned when an HTTP registry specifies an invalid format
	errInvalidFormat = errors.New("format must be 'raw' or 'tar.gz' for http registry")
	// errInvalidVersion is returned when an HTTP registry version is invalid
	errInvalidVersion = errors.New("invalid version")
)
