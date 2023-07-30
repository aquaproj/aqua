package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/util"
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
	which              which.Controller
	packageInstaller   installpackage.Installer
	executor           Executor
	fs                 afero.Fs
	policyConfigReader policy.Reader
	policyConfigFinder policy.ConfigFinder
	enabledXSysExec    bool
	requireChecksum    bool
}

func getEnabledXSysExec(osEnv osenv.OSEnv, goos string) bool {
	if goos == "windows" {
		return false
	}
	if osEnv.Getenv("AQUA_EXPERIMENTAL_X_SYS_EXEC") == "false" {
		return false
	}
	if osEnv.Getenv("AQUA_X_SYS_EXEC") == "false" {
		return false
	}
	return true
}

func New(param *config.Param, pkgInstaller installpackage.Installer, whichCtrl which.Controller, executor Executor, osEnv osenv.OSEnv, fs afero.Fs, policyConfigReader policy.Reader, policyConfigFinder policy.ConfigFinder) *Controller {
	return &Controller{
		stdin:              os.Stdin,
		stdout:             os.Stdout,
		stderr:             os.Stderr,
		packageInstaller:   pkgInstaller,
		which:              whichCtrl,
		executor:           executor,
		enabledXSysExec:    getEnabledXSysExec(osEnv, runtime.GOOS),
		fs:                 fs,
		policyConfigReader: policyConfigReader,
		policyConfigFinder: policyConfigFinder,
		requireChecksum:    param.RequireChecksum,
	}
}

func (ctrl *Controller) Exec(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string, args ...string) (gErr error) { //nolint:cyclop
	logE = logE.WithField("exe_name", exeName)
	defer func() {
		if gErr != nil {
			gErr = logerr.WithFields(gErr, logE.Data)
		}
	}()

	policyCfgs, err := ctrl.policyConfigReader.ReadFromEnv(param.PolicyConfigFilePaths)
	if err != nil {
		return fmt.Errorf("read policy files: %w", err)
	}

	globalPolicyPaths := make(map[string]struct{}, len(param.PolicyConfigFilePaths))
	for _, p := range param.PolicyConfigFilePaths {
		globalPolicyPaths[p] = struct{}{}
	}

	findResult, err := ctrl.which.Which(ctx, logE, param, exeName)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if findResult.Package == nil {
		return ctrl.execCommandWithRetry(ctx, logE, findResult.ExePath, args, nil)
	}
	logE = logE.WithFields(logrus.Fields{
		"package":         findResult.Package.Package.Name,
		"package_version": findResult.Package.Package.Version,
	})

	policyCfgs, err = ctrl.policyConfigReader.Append(logE, findResult.ConfigFilePath, policyCfgs, globalPolicyPaths)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if param.DisableLazyInstall {
		if _, err := ctrl.fs.Stat(findResult.ExePath); err != nil {
			return logerr.WithFields(errExecNotFoundDisableLazyInstall, logE.WithField("doc", "https://aquaproj.github.io/docs/reference/codes/006").Data) //nolint:wrapcheck
		}
	}
	if err := ctrl.install(ctx, logE, findResult, policyCfgs); err != nil {
		return err
	}
	var envs []string
	if findResult.Package.PackageInfo.Type == "pypi" {
		pkgPath := filepath.Dir(filepath.Dir(findResult.ExePath))
		if pythonPath, ok := os.LookupEnv("PYTHONPATH"); ok {
			envs = []string{fmt.Sprintf("PYTHONPATH=%s:%s", pkgPath, pythonPath)}
		} else {
			envs = []string{fmt.Sprintf("PYTHONPATH=%s", pkgPath)}
		}
	}
	return ctrl.execCommandWithRetry(ctx, logE, findResult.ExePath, args, envs)
}

func (ctrl *Controller) install(ctx context.Context, logE *logrus.Entry, findResult *which.FindResult, policies []*policy.Config) error {
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

	if err := ctrl.packageInstaller.InstallPackage(ctx, logE, &installpackage.ParamInstallPackage{
		Pkg:             findResult.Package,
		Checksums:       checksums,
		RequireChecksum: findResult.Config.RequireChecksum(ctrl.requireChecksum),
		PolicyConfigs:   policies,
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

func (ctrl *Controller) execCommand(ctx context.Context, exePath string, args, envs []string) (bool, error) {
	if ctrl.enabledXSysExec {
		if err := ctrl.executor.ExecXSysWithEnvs(exePath, args, envs); err != nil {
			return true, fmt.Errorf("call execve(2): %w", err)
		}
		return false, nil
	}
	if exitCode, err := ctrl.executor.ExecWithEnvs(ctx, exePath, args, envs); err != nil {
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

func (ctrl *Controller) execCommandWithRetry(ctx context.Context, logE *logrus.Entry, exePath string, args, envs []string) error {
	for i := 0; i < 10; i++ {
		logE.Debug("execute the command")
		retried, err := ctrl.execCommand(ctx, exePath, args, envs)
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
