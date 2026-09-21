package resolve

import "errors"

// errNoSupportedEnv is returned when a package resolves to nothing anywhere. A
// package that can't be installed on any environment can't be locked either, and
// returning an empty list would look like a successful resolution.
var errNoSupportedEnv = errors.New("the package supports no environment")
