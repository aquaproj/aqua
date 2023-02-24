package registry

import "errors"

var (
	errUnsupportedRegistryType = errors.New("unsupported registry type")
	errLocalRegistryNotFound   = errors.New("local registry isn't found")
	errInstallFailure          = errors.New("it failed to install some registries")
)
