package which

import (
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func New(param *config.Param, configFinder ConfigFinder, configReader reader.ConfigReader, registInstaller domain.RegistryInstaller, rt *runtime.Runtime, osEnv osenv.OSEnv, fs afero.Fs, linker domain.Linker, cosignInstaller domain.CosignInstaller) *Controller {
	return &Controller{
		stdout:            os.Stdout,
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		runtime:           rt,
		osenv:             osEnv,
		fs:                fs,
		linker:            linker,
		cosignInstaller:   cosignInstaller,
	}
}
