package exec

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type Controller struct {
	stdin            io.Reader
	stdout           io.Writer
	stderr           io.Writer
	which            which.Controller
	packageInstaller installpackage.Installer
}

func New(pkgInstaller installpackage.Installer, which which.Controller) *Controller {
	return &Controller{
		stdin:            os.Stdin,
		stdout:           os.Stdout,
		stderr:           os.Stderr,
		packageInstaller: pkgInstaller,
		which:            which,
	}
}

func (ctrl *Controller) Exec(ctx context.Context, param *config.Param, exeName string, args []string, logE *logrus.Entry) error {
	which, err := ctrl.which.Which(ctx, param, exeName, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if which.Package != nil { //nolint:nestif
		logE = logE.WithFields(logrus.Fields{
			"exe_path": which.ExePath,
			"package":  which.Package.Name,
		})
		if err := ctrl.packageInstaller.InstallPackage(ctx, which.PkgInfo, which.Package, false, logE); err != nil {
			return err //nolint:wrapcheck
		}
		for i := 0; i < 10; i++ {
			logE.Debug("check if exec file exists")
			if fi, err := os.Stat(which.ExePath); err == nil {
				if util.IsOwnerExecutable(fi.Mode()) {
					break
				}
			}
			logE.WithFields(logrus.Fields{
				"retry_count": i + 1,
			}).Debug("command isn't found. wait for lazy install")
			if err := wait(ctx, 10*time.Millisecond); err != nil { //nolint:gomnd
				return err
			}
		}
	}
	return ctrl.execCommand(ctx, which.ExePath, args, logE)
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck
	}
}

var errFailedToStartProcess = errors.New("it failed to start the process")

func (ctrl *Controller) execCommand(ctx context.Context, exePath string, args []string, logE *logrus.Entry) error {
	logE = logE.WithField("exe_path", exePath)
	for i := 0; i < 10; i++ {
		logE.Debug("execute the command")
		if err := unix.Exec(exePath, append([]string{filepath.Base(exePath)}, args...), os.Environ()); err != nil {
			logE.WithField("retry_count", i+1).Debug("the process isn't started. retry")
			if err := wait(ctx, 10*time.Millisecond); err != nil { //nolint:gomnd
				return err
			}
			continue
		}
		return nil
	}
	return errFailedToStartProcess
}
