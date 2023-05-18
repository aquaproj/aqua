package registry

import "errors"

var (
	errPkgNameIsRequired        = errors.New("package name is required")
	errRepoRequired             = errors.New("repo_owner and repo_name are required")
	errGitHubContentRequirePath = errors.New("github_content package requires path")
	errGoInstallRequirePath     = errors.New("go_install package requires path")
	errCargoRequireCrate        = errors.New("cargo package requires crate")
	errAssetRequired            = errors.New("github_release package requires asset")
	errURLRequired              = errors.New("http package requires url")
	errInvalidPackageType       = errors.New("package type is invalid")
)
