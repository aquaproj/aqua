package installpackage

import "errors"

var (
	errRegistryNotFound = errors.New("registry isn't found")
	errPkgNotFound      = errors.New("package isn't found in the registry")
	errChmod            = errors.New("add the permission to execute the command")
	errInstallFailure   = errors.New("it failed to install some packages")
)
