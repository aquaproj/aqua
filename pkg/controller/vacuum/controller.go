package vacuum

import (
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	rootDir string
	runtime *runtime.Runtime
	fs      afero.Fs
	vacuum  Vacuum
}

func New(param *config.Param, rt *runtime.Runtime, fs afero.Fs, vc Vacuum) *Controller {
	return &Controller{
		rootDir: param.RootDir,
		runtime: rt,
		fs:      fs,
		vacuum:  vc,
	}
}

type Vacuum interface {
	FindAll(logE *logrus.Entry) (map[string]time.Time, error)
	Remove(pkgPath string) error
}
