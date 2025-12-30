package list

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/spf13/afero"
)

type Controller struct {
	stdout            io.Writer
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller RegistryInstaller
	fs                afero.Fs
}

func NewController(configFinder ConfigFinder, configReader ConfigReader, registryInstaller RegistryInstaller, fs afero.Fs) *Controller {
	return &Controller{
		stdout:            os.Stdout,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registryInstaller,
		fs:                fs,
	}
}

type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
}

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}
