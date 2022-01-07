package controller

import "errors"

var (
	errPkgNameMustBeUniqueInRegistry = errors.New("the package name must be unique in the same registry")
	errInvalidRegistryType           = errors.New("registry type is invalid")
	errConfigFileNotFound            = errors.New("configuration file isn't found")
	errUnknownPkg                    = errors.New("unknown package")
	errInvalidPackageType            = errors.New("package type is invalid")
	errPathIsRequired                = errors.New("path is required for local registry")
	errRepoOwnerIsRequired           = errors.New("repo_owner is required for github_content registry")
	errRepoNameIsRequired            = errors.New("repo_name is required for github_content registry")
	errRefIsRequired                 = errors.New("ref is required for github_content registry")
	errGitHubTokenIsRequired         = errors.New("GITHUB_TOKEN is required for the type `github_release`")
	errCommandIsNotFound             = errors.New("command is not found")
	errGitHubContentMustBeFile       = errors.New("ref must be not a directory but a file")
	errUnsupportedRegistryType       = errors.New("unsupported registry type")
	errLocalRegistryNotFound         = errors.New("local registry isn't found")
	errRegistryNotFound              = errors.New("registry isn't found")
	errPkgNotFound                   = errors.New("package isn't found in the registry")
	errExePathIsDirectory            = errors.New("exe_path is directory")
	errChmod                         = errors.New("add the permission to execute the command")
	errInvalidHTTPStatusCode         = errors.New("status code >= 400")
	errInstallFailure                = errors.New("it failed to install some packages")
	errRegistryNameIsDuplicated      = errors.New("registry name is duplicated")
	errUnsupportedFileFormat         = errors.New("unsupported file format")
)
