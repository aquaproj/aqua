package vacuum

import (
	"context"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	bolt "go.etcd.io/bbolt"
)

type Controller struct {
	stdout  io.Writer
	dbMutex sync.RWMutex
	db      atomic.Pointer[bolt.DB]
	Param   *config.Param
	fs      afero.Fs
	d       *DB
}

// New initializes a Controller with the given context, parameters, and dependencies.
func New(ctx context.Context, param *config.Param, fs afero.Fs) *Controller {
	vc := &Controller{
		stdout: os.Stdout,
		Param:  param,
		fs:     fs,
		d:      NewDB(ctx, param, fs),
	}
	return vc
}

// Close closes the dependencies of the Controller.
func (vc *Controller) Close(logE *logrus.Entry) error {
	if !vc.IsVacuumEnabled(logE) {
		return nil
	}
	logE.Debug("closing vacuum controller")
	if vc.d.storeQueue != nil {
		vc.d.storeQueue.close()
	}
	return vc.d.Close()
}

func (vc *Controller) TestKeepDBOpen() error {
	return vc.d.TestKeepDBOpen()
}
