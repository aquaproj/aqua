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
	// errHostRequired is returned when a package on a forge instance lacks the instance it
	// is on.
	errHostRequired = errors.New("forgejo_release and gitea_release packages require host")
	// errForgePrivate is returned when a package on a forge instance says it is private,
	// which there is no credential to read.
	errForgePrivate = errors.New("forgejo_release and gitea_release packages don't support private")
	// errHostInvalid is returned when the instance a package is on is written as anything
	// but a host name.
	errHostInvalid = errors.New("host must be a host name, with no scheme, userinfo, port or path")
	// errForgeFileSource is returned when a package reads a file beside its asset as a file
	// on a forge instance that isn't the instance it is on itself.
	errForgeFileSource = errors.New("a forgejo_release or gitea_release checksum file or signature requires a package of the same type")
	// errGitHubFileSource is returned when a package on a forge instance reads a file beside
	// its asset from github.com, where the same repository is somebody else's.
	errGitHubFileSource = errors.New("a github_release checksum file or signature on a forgejo_release or gitea_release package must name the repository it is in")
	// errForgeGitHubVerification is returned when a package on a forge instance asks for a
	// verification that only GitHub can answer.
	errForgeGitHubVerification = errors.New("forgejo_release and gitea_release packages don't support slsa_provenance or github_artifact_attestations")
	// errInvalidPackageType is returned when a package has an unrecognized type.
	errInvalidPackageType = errors.New("package type is invalid")
)
