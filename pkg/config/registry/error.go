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
	// errAssetRequired is returned when a release package lacks an asset specification.
	errAssetRequired = errors.New("github_release and forgejo_release packages require asset")
	// errURLRequired is returned when an http package lacks a URL.
	errURLRequired = errors.New("http package requires url")
	// errHostRequired is returned when a forgejo_release package lacks the instance it is on.
	errHostRequired = errors.New("forgejo_release package requires host")
	// errForgejoPrivate is returned when a forgejo_release package says it is private,
	// which there is no credential to read.
	errForgejoPrivate = errors.New("forgejo_release package doesn't support private")
	// errForgejoGitHubVerification is returned when a forgejo_release package asks for a
	// verification that only GitHub can answer.
	errForgejoGitHubVerification = errors.New("forgejo_release package doesn't support slsa_provenance or github_artifact_attestations")
	// errInvalidPackageType is returned when a package has an unrecognized type.
	errInvalidPackageType = errors.New("package type is invalid")
)
