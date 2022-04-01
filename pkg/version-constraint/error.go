package constraint

import "errors"

var (
	errPackageConditionMustBeBoolean   = errors.New("PackageCondition must be a boolean")
	errVersionConstraintsMustBeBoolean = errors.New("VersionConstraints must be a boolean")
	errVersionFilterMustBeBoolean      = errors.New("VersionFilter must be a boolean")
)
