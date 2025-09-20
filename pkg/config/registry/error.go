package registry

import "errors"

// Package validation errors returned by PackageInfo.Validate().
var (
	// errPkgNameIsRequired is returned when a package has no name or derivable name.
	errPkgNameIsRequired = errors.New("package name is required")
	// errRepoRequired is returned when a package type requires repository information but it's missing.
	errRepoRequired = errors.New("repo_owner and repo_name are required")
	// errGitHubContentRequirePath is returned when a github_content package lacks a path.
	errGitHubContentRequirePath = errors.New("github_content package requires path")
	// errGoInstallRequirePath is returned when a go_install package lacks a path.
	errGoInstallRequirePath = errors.New("go_install package requires path")
	// errCargoRequireCrate is returned when a cargo package lacks a crate name.
	errCargoRequireCrate = errors.New("cargo package requires crate")
	// errAssetRequired is returned when a github_release package lacks an asset specification.
	errAssetRequired = errors.New("github_release package requires asset")
	// errURLRequired is returned when an http package lacks a URL.
	errURLRequired = errors.New("http package requires url")
	// errInvalidPackageType is returned when a package has an unrecognized type.
	errInvalidPackageType = errors.New("package type is invalid")
)
