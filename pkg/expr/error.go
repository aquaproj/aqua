package expr

import "errors"

var (
	errMustBeBoolean = errors.New("the evaluation result must be a boolean")
	errMustBeString  = errors.New("the evaluation result must be a string")
)
