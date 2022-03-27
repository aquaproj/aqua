package validate

import "errors"

var (
	errPkgNameMustBeUniqueInRegistry = errors.New("the package name must be unique in the same registry")
	ErrRegistryNameIsDuplicated      = errors.New("registry name is duplicated")
)
