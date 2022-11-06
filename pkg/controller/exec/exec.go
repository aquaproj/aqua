package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/policy"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	stdin              io.Reader
	stdout             io.Writer
	stderr             io.Writer
	which              domain.WhichController
	packageInstaller   domain.PackageInstaller
	executor           Executor
	enabledXSysExec    bool
	fs                 afero.Fs
	policyConfigReader domain.PolicyConfigReader
	policyChecker      domain.PolicyChecker
}

type Executor interface {
	Exec(ctx context.Context, exePath string, args []string) (int, error)
	ExecXSys(exePath string, args []string) error
}

func New(pkgInstaller domain.PackageInstaller, whichCtrl domain.WhichController, executor Executor, osEnv osenv.OSEnv, fs afero.Fs, policyConfigReader domain.PolicyConfigReader, policyChecker domain.PolicyChecker) *Controller {
	return &Controller{
		stdin:              os.Stdin,
		stdout:             os.Stdout,
		stderr:             os.Stderr,
		packageInstaller:   pkgInstaller,
		which:              whichCtrl,
		executor:           executor,
		enabledXSysExec:    osEnv.Getenv("AQUA_EXPERIMENTAL_X_SYS_EXEC") == "true",
		fs:                 fs,
		policyConfigReader: policyConfigReader,
		policyChecker:      policyChecker,
	}
}

func (ctrl *Controller) Exec(ctx context.Context, param *config.Param, exeName string, args []string, logE *logrus.Entry) error {
	logE = logE.WithField("exe_name", exeName)
	findResult, err := ctrl.which.Which(ctx, param, exeName, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if findResult.Package != nil {
		logE = logE.WithFields(logrus.Fields{
			"package":         findResult.Package.Package.Name,
			"package_version": findResult.Package.Package.Version,
		})
		if err := ctrl.validate(findResult.Package, filepath.Dir(findResult.ConfigFilePath), param.PolicyConfigFilePath); err != nil {
			return err
		}
		if err := ctrl.install(ctx, logE, findResult); err != nil {
			return logerr.WithFields(err, logE.Data) //nolint:wrapcheck
		}
	}
	return logerr.WithFields(ctrl.execCommandWithRetry(ctx, findResult.ExePath, args, logE), logE.Data) //nolint:wrapcheck
}

func (ctrl *Controller) validate(pkg *config.Package, cfgDir, policyConfigFilePath string) error {
	policyCfg := &policy.Config{}
	if policyConfigFilePath != "" {
		if err := ctrl.policyConfigReader.Read(policyConfigFilePath, policyCfg); err != nil {
			return fmt.Errorf("read the policy config file: %w", err)
		}
		if err := policyCfg.Init(); err != nil {
			return fmt.Errorf("parse the policy file: %w", err)
		}
		if err := ctrl.policyChecker.ValidatePackage(&config.ParamValidatePackage{
			Pkg:           pkg,
			PolicyConfig:  policyCfg,
			ConfigFileDir: cfgDir,
			PolicyFileDir: filepath.Dir(policyConfigFilePath),
		}); err != nil {
			return fmt.Errorf("validate the installed package for security: %w", err)
		}
	}
	return nil
}

func (ctrl *Controller) install(ctx context.Context, logE *logrus.Entry, findResult *domain.FindResult) error {
	logE = logE.WithField("exe_path", findResult.ExePath)

	var checksums *checksum.Checksums
	if findResult.Config.ChecksumEnabled() {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(ctrl.fs, findResult.ConfigFilePath)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := checksums.ReadFile(ctrl.fs, checksumFilePath); err != nil {
			return fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(ctrl.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}

	if err := ctrl.packageInstaller.InstallPackage(ctx, logE, &domain.ParamInstallPackage{
		Pkg:             findResult.Package,
		Checksums:       checksums,
		RequireChecksum: findResult.Config.RequireChecksum(),
	}); err != nil {
		return err //nolint:wrapcheck
	}
	for i := 0; i < 10; i++ {
		logE.Debug("check if exec file exists")
		if fi, err := ctrl.fs.Stat(findResult.ExePath); err == nil {
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
	return nil
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
