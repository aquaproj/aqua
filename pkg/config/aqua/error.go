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
)
