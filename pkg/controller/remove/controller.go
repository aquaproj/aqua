package remove

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	rgst "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

type Controller struct {
	rootDir           string
	fs                afero.Fs
	runtime           *runtime.Runtime
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller rgst.Installer
	fuzzyFinder       FuzzyFinder
	which             which.Controller
}

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}

type FuzzyFinder interface {
	FindMulti(pkgs []*fuzzyfinder.Item, hasPreview bool) ([]int, error)
}

func New(param *config.Param, fs afero.Fs, rt *runtime.Runtime, configFinder ConfigFinder, configReader ConfigReader, registryInstaller rgst.Installer, fuzzyFinder FuzzyFinder, whichController which.Controller) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		fs:                fs,
		runtime:           rt,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registryInstaller,
		fuzzyFinder:       fuzzyFinder,
		which:             whichController,
	}
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
}
