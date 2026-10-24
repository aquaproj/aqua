package versiongetter

import "errors"

var (
	// errForgeHostRequired is returned when a package on a forge instance doesn't say
	// which instance it is on. Validation refuses such a definition, so this is the case
	// of a package read some other way.
	errForgeHostRequired = errors.New("forgejo_release, gitea_release and gitlab_release packages require host")
	// errForgeTagHostRequired is returned when a package reading a forge's tags doesn't
	// say which instance they are on.
	errForgeTagHostRequired = errors.New("version_source forgejo_tag and gitea_tag require host")
	// errForgeClientMissing is returned when nothing reads the forge the package is on,
	// which is a getter built without that client rather than a definition being wrong.
	errForgeClientMissing = errors.New("nothing reads the versions of this package's forge")
)
