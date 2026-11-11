package migrate

import "errors"

var (
	// errNoLockFile is returned when there is no lock file to take the checksums'
	// place.
	errNoLockFile = errors.New("there is no lock file, so nothing would verify the packages; run 'aqua lock update' first")
	// errOnlySection is returned for a configuration that is nothing but a checksum
	// section, which is not something to write back empty.
	errOnlySection = errors.New("the configuration says nothing but the checksum settings")
	// errNotLocked is returned when the lock file doesn't cover every package.
	errNotLocked = errors.New("the lock file doesn't hold every package the configuration asks for; run 'aqua lock update' first")
)
