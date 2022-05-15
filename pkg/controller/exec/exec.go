package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

type Controller struct {
	stdin            io.Reader
	stdout           io.Writer
	stderr           io.Writer
	which            which.Controller
	packageInstaller installpackage.Installer
	executor         exec.Executor
	enabledXSysExec  bool
	fs               afero.Fs
}

func New(pkgInstaller installpackage.Installer, which which.Controller, executor exec.Executor, osEnv osenv.OSEnv, fs afero.Fs) *Controller {
	return &Controller{
		stdin:            os.Stdin,
		stdout:           os.Stdout,
		stderr:           os.Stderr,
		packageInstaller: pkgInstaller,
		which:            which,
		executor:         executor,
		enabledXSysExec:  osEnv.Getenv("AQUA_EXPERIMENTAL_X_SYS_EXEC") == "true",
		fs:               fs,
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
			if fi, err := ctrl.fs.Stat(which.ExePath); err == nil {
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
	return ctrl.execCommandWithRetry(ctx, which.ExePath, args, logE)
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

func (ctrl *Controller) execCommand(ctx context.Context, exePath string, args []string) (bool, error) {
	if ctrl.enabledXSysExec {
		if err := ctrl.executor.ExecXSys(exePath, args); err != nil {
			return true, fmt.Errorf("call execve(2): %w", err)
		}
		return false, nil
	}
	if exitCode, err := ctrl.executor.Exec(ctx, exePath, args); err != nil {
		// https://pkg.go.dev/os#ProcessState.ExitCode
		// > ExitCode returns the exit code of the exited process,
		// > or -1 if the process hasn't exited or was terminated by a signal.
		if exitCode == -1 && ctx.Err() == nil {
			return true, fmt.Errorf("execute a command: %w", err)
		}
		return false, ecerror.Wrap(err, exitCode)
	}
	return false, nil
}

func (ctrl *Controller) execCommandWithRetry(ctx context.Context, exePath string, args []string, logE *logrus.Entry) error {
	logE = logE.WithField("exe_path", exePath)
	for i := 0; i < 10; i++ {
		logE.Debug("execute the command")
		retried, err := ctrl.execCommand(ctx, exePath, args)
		if !retried {
			return err
		}
		logE.WithError(err).WithField("retry_count", i+1).Debug("the process isn't started. retry")
		if err := wait(ctx, 10*time.Millisecond); err != nil { //nolint:gomnd
			return err
		}
	}
	return errFailedToStartProcess
}
