package which

import "errors"

var (
	ErrCommandIsNotFound = errors.New("command is not found")
	errVersionIsRequired = errors.New("version is required")
)
