package download

import "errors"

var (
	errGitHubContentMustBeFile = errors.New("ref must be not a directory but a file")
	errInvalidHTTPStatusCode   = errors.New("status code >= 400")
)
