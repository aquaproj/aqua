package apperr

import "errors"

var (
	ErrRepoRequired      = errors.New("repo_owner and repo_name are required")
	ErrPkgNameIsRequired = errors.New("package name is required")
)
