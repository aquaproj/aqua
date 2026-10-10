package versiongetter

import "errors"

var (
	// errForgeHostRequired is returned when a package on a forge instance doesn't say
	// which instance it is on. Validation refuses such a definition, so this is the case
	// of a package read some other way.
	errForgeHostRequired = errors.New("forgejo_release, gitea_release and gitlab_release packages require host")
	// errForgeClientMissing is returned when nothing reads the forge the package is on,
	// which is a getter built without that client rather than a definition being wrong.
	errForgeClientMissing = errors.New("nothing reads the releases of this package's forge")
)
