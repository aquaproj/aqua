package download

import "errors"

var (
	errInvalidPackageType      = errors.New("package type is invalid")
	errGitHubTokenIsRequired   = errors.New("GITHUB_TOKEN is required for the type `github_release`")
	errGitHubContentMustBeFile = errors.New("path must be not a directory but a file")
	errInvalidHTTPStatusCode   = errors.New("status code >= 400")
)
