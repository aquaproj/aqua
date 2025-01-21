package vacuum

import (
	"context"
	"io"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/spf13/afero"
)

type Controller struct {
	stdout io.Writer
	Param  *config.Param
	fs     afero.Fs
	db     *DB
}

// New initializes a Controller with the given context, parameters, and dependencies.
func New(ctx context.Context, param *config.Param, fs afero.Fs) *Controller {
	vc := &Controller{
		stdout: os.Stdout,
		Param:  param,
		fs:     fs,
		db:     NewDB(ctx, param, fs),
	}
	return vc
}
