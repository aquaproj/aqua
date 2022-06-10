package which

import "errors"

var (
	errCommandIsNotFound = errors.New("command is not found")
	errVersionIsRequired = errors.New("version is required")
)
