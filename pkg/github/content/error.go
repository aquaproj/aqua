package content

import "errors"

var errGitHubContentMustBeFile = errors.New("ref must be not a directory but a file")
