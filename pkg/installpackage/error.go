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
