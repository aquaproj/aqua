package installpackage

import "errors"

var (
	errExePathIsDirectory    = errors.New("exe_path is directory")
	errChmod                 = errors.New("add the permission to execute the command")
	errInstallFailure        = errors.New("it failed to install some packages")
	errGoInstallForbidLatest = errors.New(`the version "latest" is forbidden. Please specify Git tag or commit sha`)
	errInvalidChecksum       = errors.New("checksum is invalid")
	errChecksumIsRequired    = errors.New("checksum is required")
	errNoAsset               = errors.New("no asset is released for this version")
)

var (
	errNoChecksumInLockFile = errors.New("the lock file entry has no checksum")
	errPkgNameIsEmpty       = errors.New("the package name is empty")
	errPkgVersionIsEmpty    = errors.New("the package version is empty")
	// errNotInLockFile is what a package missing from an existing lock file gets.
	// The lock file is what aqua installs from, so the fix is to put the package in
	// it, not to work around it.
	errNotInLockFile = errors.New(`the package isn't in the lock file. Please run "aqua lock update"`)
	errLockFile      = errors.New("some packages can't be installed from the lock file")
)
