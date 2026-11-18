package lockupdate

import "errors"

var (
	errUpdateLockFile = errors.New("update the lock file")
	errNoChecksum     = errors.New("the checksum of the package isn't found")
)
