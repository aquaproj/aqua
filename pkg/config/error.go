package config

import "errors"

var (
	errPkgNameMustBeUniqueInRegistry = errors.New("the package name must be unique in the same registry")
	errInvalidRegistryType           = errors.New("registry type is invalid")
	errPathIsRequired                = errors.New("path is required for local registry")
	errRepoOwnerIsRequired           = errors.New("repo_owner is required")
	errRepoNameIsRequired            = errors.New("repo_name is required")
	errRefIsRequired                 = errors.New("ref is required for github_content registry")
)
