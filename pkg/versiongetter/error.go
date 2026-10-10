package versiongetter

import "errors"

// errForgeHostRequired is returned when a package on a forge instance doesn't say which
// instance it is on. Validation refuses such a definition, so this is the case of a
// package read some other way.
var errForgeHostRequired = errors.New("forgejo_release and gitea_release packages require host")
