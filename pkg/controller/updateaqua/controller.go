package updateaqua

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

type Controller struct {
	rootDir   string
	fs        afero.Fs
	runtime   *runtime.Runtime
	github    RepositoriesService
	installer AquaInstaller
}

func New(param *config.Param, fs afero.Fs, rt *runtime.Runtime, gh RepositoriesService, installer AquaInstaller) *Controller {
	return &Controller{
		rootDir:   param.RootDir,
		fs:        fs,
		runtime:   rt,
		github:    gh,
		installer: installer,
	}
}
