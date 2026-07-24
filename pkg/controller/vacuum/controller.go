package vacuum

import (
	"log/slog"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type Controller struct {
	rootDir string
	runtime *runtime.Runtime
	vacuum  Vacuum
}

func New(param *config.Param, rt *runtime.Runtime, vc Vacuum) *Controller {
	return &Controller{
		rootDir: param.RootDir,
		runtime: rt,
		vacuum:  vc,
	}
}

type Vacuum interface {
	FindAll(logger *slog.Logger) (map[string]time.Time, error)
	Remove(pkgPath string) error
}
