package which

import "errors"

var (
	ErrCommandIsNotFound = errors.New("command is not found")
	errVersionIsRequired = errors.New("version is required")
)

// errNotInLockFile is what a package missing from an existing lock file gets. The
// lock file is what aqua installs from, so a command it doesn't describe is one aqua
// would refuse to install.
var errNotInLockFile = errors.New(`the package isn't in the lock file. Please run "aqua lock update"`)
