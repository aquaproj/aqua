package registry

import "errors"

var (
	errUnsupportedRegistryType       = errors.New("unsupported registry type")
	errLocalRegistryNotFound         = errors.New("local registry isn't found")
	errInstallFailure                = errors.New("it failed to install some registries")
	errInvalidRegistryFormat         = errors.New("invalid registry format")
	errRegistryFileNotFoundInArchive = errors.New("registry file not found in archive")
)
