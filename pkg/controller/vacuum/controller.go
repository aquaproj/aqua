package vacuum

import (
	"log/slog"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
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
	FindAll(logger *slog.Logger) (map[string]time.Time, error)
	Remove(pkgPath string) error
}
