package initialize

import (
	"context"
	"log/slog"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type Controller struct {
	rootDir           string
	runtime           *runtime.Runtime
	vacuum            Vacuum
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller RegistryInstaller
	g2                lockresolve.G2
}

func New(param *config.Param, rt *runtime.Runtime, vc Vacuum, configFinder ConfigFinder, configReader ConfigReader, registryInstaller RegistryInstaller, g2 lockresolve.G2) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		runtime:           rt,
		vacuum:            vc,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registryInstaller,
		g2:                g2,
	}
}

type Vacuum interface {
	Create(pkgPath string, timestamp time.Time) error
}

type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

// RegistryInstaller installs the registries the packages' paths are read from: every
// registry for a configuration without a lock file, or the one registry of each package for
// a configuration with one.
type RegistryInstaller interface {
	lockresolve.RegistryInstaller
	InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}
