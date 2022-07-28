package which

import (
	"context"
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	cfgRegistry "github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

type Controller interface {
	Which(ctx context.Context, param *config.Param, exeName string, logE *logrus.Entry) (*Which, error)
}

func New(param *config.Param, configFinder ConfigFinder, configReader domain.ConfigReader, registInstaller domain.RegistryInstaller, rt *runtime.Runtime, osEnv osenv.OSEnv, fs afero.Fs, linker link.Linker) Controller {
	return &controller{
		stdout:            os.Stdout,
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		runtime:           rt,
		osenv:             osEnv,
		fs:                fs,
		linker:            linker,
	}
}

type Which struct {
	Package *config.Package
	File    *cfgRegistry.File
	ExePath string
}
