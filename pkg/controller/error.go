package controller

import "errors"

var (
	errPkgNameMustBeUniqueInRegistry = errors.New("the package name must be unique in the same registry")
	errUnknownPkg                    = errors.New("unknown package")
	errCommandIsNotFound             = errors.New("command is not found")
	errUnsupportedRegistryType       = errors.New("unsupported registry type")
	errLocalRegistryNotFound         = errors.New("local registry isn't found")
	errRegistryNotFound              = errors.New("registry isn't found")
	errPkgNotFound                   = errors.New("package isn't found in the registry")
	errExePathIsDirectory            = errors.New("exe_path is directory")
	errChmod                         = errors.New("add the permission to execute the command")
	errInstallFailure                = errors.New("it failed to install some packages")
	errRegistryNameIsDuplicated      = errors.New("registry name is duplicated")
)
