package updateaqua

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type Controller struct {
	rootDir   string
	runtime   *runtime.Runtime
	github    RepositoriesService
	installer AquaInstaller
}

func New(param *config.Param, rt *runtime.Runtime, gh RepositoriesService, installer AquaInstaller) *Controller {
	return &Controller{
		rootDir:   param.RootDir,
		runtime:   rt,
		github:    gh,
		installer: installer,
	}
}
