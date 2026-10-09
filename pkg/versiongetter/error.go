package versiongetter

import "errors"

// errForgejoHostRequired is returned when a forgejo_release package doesn't say which
// instance it is on. Validation refuses such a definition, so this is the case of a
// package read some other way.
var errForgejoHostRequired = errors.New("forgejo_release package requires host")
