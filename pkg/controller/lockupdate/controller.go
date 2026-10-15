// Package lockupdate implements the aqua lock update command.
//
// The command builds aqua-lock.json from aqua.yaml. aqua.yaml says which package and
// which version; the lock file says what that version actually is on every
// environment, down to the checksum. Installing then needs no registry, which is what
// makes the lock file the thing aqua trusts rather than a cache of something else.
package lockupdate

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
)

// ConfigFinder lists the aqua.yaml files that apply to a directory.
type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

// ConfigReader reads an aqua.yaml, resolving its imports.
type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
}

// Resolver turns a package version into lock file entries, one per environment.
//
// This is the arm for the standard registry, which aqua-registry-g2 mirrors: it reads
// a generated registry.json, falling back to the local cache first. A package from any
// other registry has no branch there and is resolved from its registry instead.
type Resolver interface {
	Resolve(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]*lockfile.Package, error)
}

// ChecksumGetter reads the checksum of every environment a package supports, after
// verifying whatever signature the package publishes over it. The lock file stands in
// for those signatures once it is written, so this is the only place they are seen.
type ChecksumGetter interface {
	GetVerified(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, supportedEnvs []string) error
}

// RegistryInstaller installs the registries a configuration declares.
type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}

// Controller updates lock files.
type Controller struct {
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller RegistryInstaller
	checksumGetter    ChecksumGetter
	g2                Resolver
}

// New creates a Controller.
func New(configFinder ConfigFinder, configReader ConfigReader, registryInstaller RegistryInstaller, checksumGetter ChecksumGetter, g2 Resolver) *Controller {
	return &Controller{
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registryInstaller,
		checksumGetter:    checksumGetter,
		g2:                g2,
	}
}
