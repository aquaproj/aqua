package constraint

import "errors"

var (
	errVersionConstraintsMustBeBoolean = errors.New("VersionConstraints must be a boolean")
	errVersionFilterMustBeBoolean      = errors.New("VersionFilter must be a boolean")
	errMustBeBoolean                   = errors.New("the evaluation result must be a boolean")
)
