package exec

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) Exec(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string, args ...string) (gErr error) {
	logE = logE.WithField("exe_name", exeName)
	defer func() {
		if gErr != nil {
			gErr = logerr.WithFields(gErr, logE.Data)
		}
	}()

	policyCfgs, err := c.policyConfigReader.Read(param.PolicyConfigFilePaths)
	if err != nil {
		return fmt.Errorf("read policy files: %w", err)
	}

	globalPolicyPaths := make(map[string]struct{}, len(param.PolicyConfigFilePaths))
	for _, p := range param.PolicyConfigFilePaths {
		globalPolicyPaths[p] = struct{}{}
	}

	findResult, err := c.which.Which(ctx, logE, param, exeName)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if findResult.Package == nil {
		return c.execCommandWithRetry(ctx, logE, findResult.ExePath, args...)
	}

	logE = logE.WithFields(logrus.Fields{
		"package_name":    findResult.Package.Package.Name,
		"package_version": findResult.Package.Package.Version,
	})

	policyCfgs, err = c.policyConfigReader.Append(logE, findResult.ConfigFilePath, policyCfgs, globalPolicyPaths)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if param.DisableLazyInstall {
		if _, err := c.fs.Stat(findResult.ExePath); err != nil {
			return logerr.WithFields(errExecNotFoundDisableLazyInstall, logE.WithField("doc", "https://aquaproj.github.io/docs/reference/codes/006").Data) //nolint:wrapcheck
		}
	}
	if err := c.install(ctx, logE, findResult, policyCfgs); err != nil {
		return err
	}
	return c.execCommandWithRetry(ctx, logE, findResult.ExePath, args...)
}

func (c *Controller) install(ctx context.Context, logE *logrus.Entry, findResult *which.FindResult, policies []*policy.Config) error {
	var checksums *checksum.Checksums
	if findResult.Config.ChecksumEnabled() {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(c.fs, findResult.ConfigFilePath)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := checksums.ReadFile(c.fs, checksumFilePath); err != nil {
			return fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(c.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}

	if err := c.packageInstaller.InstallPackage(ctx, logE, &installpackage.ParamInstallPackage{
		Pkg:             findResult.Package,
		Checksums:       checksums,
		RequireChecksum: findResult.Config.RequireChecksum(c.requireChecksum),
		PolicyConfigs:   policies,
	}); err != nil {
		return fmt.Errorf("install the package: %w", err)
	}
	for i := 0; i < 10; i++ {
		logE.Debug("check if exec file exists")
		if fi, err := c.fs.Stat(findResult.ExePath); err == nil {
			if osfile.IsOwnerExecutable(fi.Mode()) {
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

func (c *Controller) execCommand(ctx context.Context, exePath string, args ...string) (bool, error) {
	if c.enabledXSysExec {
		if err := c.executor.ExecXSys(exePath, args...); err != nil {
			return true, fmt.Errorf("call execve(2): %w", err)
		}
		return false, nil
	}
	if exitCode, err := c.executor.Exec(ctx, exePath, args...); err != nil {
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

func (c *Controller) execCommandWithRetry(ctx context.Context, logE *logrus.Entry, exePath string, args ...string) error {
	for i := 0; i < 10; i++ {
		logE.Debug("execute the command")
		retried, err := c.execCommand(ctx, exePath, args...)
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
