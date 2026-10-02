package resolve

import "errors"

// errNoSupportedEnv is returned when a package resolves to nothing anywhere. A
// package that can't be installed on any environment can't be locked either, and
// returning an empty list would look like a successful resolution.
var errNoSupportedEnv = errors.New("the package supports no environment")

// errVarHasNoValue is returned for a variable the definition declares but nothing
// fills. Static resolution has no one to ask for a value, so the variable would be
// rendered as "<no value>" into whatever the definition used it for.
var errVarHasNoValue = errors.New("a package variable has no value")
