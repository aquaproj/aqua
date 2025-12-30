package exec

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) Exec(ctx context.Context, logger *slog.Logger, param *config.Param, exeName string, args ...string) (gErr error) { //nolint:cyclop
	logger = logger.With("exe_name", exeName)
	defer func() {
		if gErr != nil {
			gErr = slogerr.With(gErr)
		}
	}()

	policyCfgs, err := c.policyReader.Read(param.PolicyConfigFilePaths)
	if err != nil {
		return fmt.Errorf("read policy files: %w", err)
	}

	globalPolicyPaths := make(map[string]struct{}, len(param.PolicyConfigFilePaths))
	for _, p := range param.PolicyConfigFilePaths {
		globalPolicyPaths[p] = struct{}{}
	}

	findResult, err := c.which.Which(ctx, logger, param, exeName)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if findResult.Package == nil {
		return c.execCommandWithRetry(ctx, logger, findResult.ExePath, exeName, args...)
	}

	logger = logger.With(
		"package_name", findResult.Package.Package.Name,
		"package_version", findResult.Package.Package.Version,
	)

	policyCfgs, err = c.policyReader.Append(logger, findResult.ConfigFilePath, policyCfgs, globalPolicyPaths)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if param.DisableLazyInstall {
		if _, err := c.fs.Stat(findResult.ExePath); err != nil {
			return slogerr.With(errExecNotFoundDisableLazyInstall, "doc", "https://aquaproj.github.io/docs/reference/codes/006") //nolint:wrapcheck
		}
	}
	if err := c.install(ctx, logger, findResult, policyCfgs, param); err != nil {
		return err
	}

	if err := c.updateTimestamp(findResult.Package); err != nil {
		slogerr.WithError(logger, err).Warn("update the last used datetime")
	}

	return c.execCommandWithRetry(ctx, logger, findResult.ExePath, exeName, args...)
}

func (c *Controller) updateTimestamp(pkg *config.Package) error {
	pkgPath, err := pkg.PkgPath(runtime.New())
	if err != nil {
		return fmt.Errorf("get a package path: %w", err)
	}
	if err := c.vacuum.Update(pkgPath, time.Now()); err != nil {
		return fmt.Errorf("update the last used datetime: %w", err)
	}
	return nil
}

func (c *Controller) install(ctx context.Context, logger *slog.Logger, findResult *which.FindResult, policies []*policy.Config, param *config.Param) error {
	checksums, updateChecksum, err := checksum.Open(
		logger, c.fs, findResult.ConfigFilePath,
		param.ChecksumEnabled(findResult.Config))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	if err := c.packageInstaller.InstallPackage(ctx, logger, &installpackage.ParamInstallPackage{
		Pkg:             findResult.Package,
		Checksums:       checksums,
		RequireChecksum: findResult.Config.RequireChecksum(param.EnforceRequireChecksum, param.RequireChecksum),
		PolicyConfigs:   policies,
		DisablePolicy:   param.DisablePolicy,
	}); err != nil {
		return fmt.Errorf("install the package: %w", err)
	}
	for i := range 10 {
		logger.Debug("check if exec file exists")
		if fi, err := c.fs.Stat(findResult.ExePath); err == nil {
			if osfile.IsOwnerExecutable(fi.Mode()) {
				break
			}
		}
		logger.Debug("command isn't found. wait for lazy install",
			"retry_count", i+1)
		if err := wait(ctx, 10*time.Millisecond); err != nil { //nolint:mnd
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

func (c *Controller) execCommand(ctx context.Context, exePath, name string, args ...string) (bool, error) {
	if c.enabledXSysExec {
		if err := c.executor.ExecXSys(exePath, name, args...); err != nil {
			return true, fmt.Errorf("call execve(2): %w", err)
		}
		return false, nil
	}
	cmd := osexec.Command(ctx, exePath, args...)
	cmd.Args[0] = name
	if exitCode, err := c.executor.Exec(cmd); err != nil {
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

func (c *Controller) execCommandWithRetry(ctx context.Context, logger *slog.Logger, exePath, name string, args ...string) error {
	for i := range 10 {
		logger.Debug("execute the command")
		retried, err := c.execCommand(ctx, exePath, name, args...)
		if !retried {
			return err
		}
		slogerr.WithError(logger, err).Debug("the process isn't started. retry", "retry_count", i+1)
		if err := wait(ctx, 10*time.Millisecond); err != nil { //nolint:mnd
			return err
		}
	}
	return errFailedToStartProcess
}
