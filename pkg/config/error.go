package config

import "errors"

var (
	errPkgNameMustBeUniqueInRegistry = errors.New("the package name must be unique in the same registry")
	errPkgNameIsRequired             = errors.New("package name is required")
	errRepoRequired                  = errors.New("repo_owner and repo_name are required")
	errGitHubContentRequirePath      = errors.New("github_content package requires path")
	errAssetRequired                 = errors.New("github_release package requires asset")
	errURLRequired                   = errors.New("http package requires url")
	errInvalidPackageType            = errors.New("package type is invalid")

	errInvalidRegistryType = errors.New("registry type is invalid")
	errPathIsRequired      = errors.New("path is required for local registry")
	errRepoOwnerIsRequired = errors.New("repo_owner is required")
	errRepoNameIsRequired  = errors.New("repo_name is required")
	errRefIsRequired       = errors.New("ref is required for github_content registry")
)
