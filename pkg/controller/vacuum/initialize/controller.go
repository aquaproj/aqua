package initialize

import (
	"context"
	"log/slog"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

type Controller struct {
	rootDir           string
	runtime           *runtime.Runtime
	fs                afero.Fs
	vacuum            Vacuum
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller RegistryInstaller
}

func New(param *config.Param, rt *runtime.Runtime, fs afero.Fs, vc Vacuum, configFinder ConfigFinder, configReader ConfigReader, registryInstaller RegistryInstaller) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		runtime:           rt,
		fs:                fs,
		vacuum:            vc,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registryInstaller,
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

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}
