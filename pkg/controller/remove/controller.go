package remove

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

type Controller struct {
	rootDir           string
	fs                afero.Fs
	runtime           *runtime.Runtime
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller RegistryInstaller
	fuzzyFinder       FuzzyFinder
	which             WhichController
	mode              *config.RemoveMode
	vacuum            Vacuum
}

type WhichController interface {
	Which(ctx context.Context, logger *slog.Logger, param *config.Param, exeName string) (*which.FindResult, error)
}

type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
}

type FuzzyFinder interface {
	FindMulti(pkgs []*fuzzyfinder.Item, hasPreview bool) ([]int, error)
}

func New(param *config.Param, target *config.RemoveMode, fs afero.Fs, rt *runtime.Runtime, configFinder ConfigFinder, configReader ConfigReader, registryInstaller RegistryInstaller, fuzzyFinder FuzzyFinder, whichController WhichController, vacuum Vacuum) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		fs:                fs,
		runtime:           rt,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registryInstaller,
		fuzzyFinder:       fuzzyFinder,
		which:             whichController,
		mode:              target,
		vacuum:            vacuum,
	}
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
}

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}

type Vacuum interface {
	Remove(pkgPath string) error
}
