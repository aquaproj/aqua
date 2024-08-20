package setshell

import (
	"io"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

type Controller struct {
	rootDir string
	fs      afero.Fs
	runtime *runtime.Runtime
	stdout  io.Writer
}

func New(param *config.Param, fs afero.Fs, rt *runtime.Runtime, stdout io.Writer) *Controller {
	return &Controller{
		rootDir: param.RootDir,
		fs:      fs,
		runtime: rt,
		stdout:  stdout,
	}
}
