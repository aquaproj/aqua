package lockupdate

import "errors"

var errUpdateLockFile = errors.New("update the lock file")

var (
	errRegistryNotFound      = errors.New("the registry isn't found")
	errPkgNotFoundInRegistry = errors.New("the package isn't found in the registry")
	errNoChecksum            = errors.New("the checksum of the package isn't found")
)
