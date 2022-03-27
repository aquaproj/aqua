package controller

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/aquaproj/aqua/pkg/validate"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) Exec(ctx context.Context, param *config.Param, exeName string, args []string) error {
	which, err := ctrl.which(ctx, param, exeName)
	if err != nil {
		return err
	}
	if which.Package != nil { //nolint:nestif
		logE := ctrl.logE().WithFields(logrus.Fields{
			"exe_path": which.ExePath,
			"package":  which.Package.Name,
		})
		if err := ctrl.PackageInstaller.InstallPackage(ctx, which.PkgInfo, which.Package, false); err != nil {
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
	return ctrl.execCommand(ctx, which.ExePath, args)
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

func (ctrl *Controller) findExecFileFromPkg(registries map[string]*config.RegistryContent, exeName string, pkg *config.Package) (*config.PackageInfo, *config.File) {
	logE := ctrl.logE().WithFields(logrus.Fields{
		"registry_name": pkg.Registry,
		"package_name":  pkg.Name,
	})
	registry, ok := registries[pkg.Registry]
	if !ok {
		logE.Warn("registry isn't found")
		return nil, nil
	}

	m, err := registry.PackageInfos.ToMap()
	if err != nil {
		logerr.WithError(logE, err).Warn("registry is invalid")
		return nil, nil
	}

	pkgInfo, ok := m[pkg.Name]
	if !ok {
		logE.Warn("package isn't found")
		return nil, nil
	}

	pkgInfo, err = pkgInfo.SetVersion(pkg.Version)
	if err != nil {
		logerr.WithError(logE, err).Warn("version constraint is invalid")
		return nil, nil
	}

	if pkgInfo.SupportedIf != nil {
		supported, err := pkgInfo.SupportedIf.Check()
		if err != nil {
			logerr.WithError(logE, err).WithField("supported_if", pkgInfo.SupportedIf.Raw()).Error("check if the package is supported")
			return nil, nil
		}
		if !supported {
			logE.WithField("supported_if", pkgInfo.SupportedIf.Raw()).Debug("the package isn't supported on this environment")
			return nil, nil
		}
	}

	for _, file := range pkgInfo.GetFiles() {
		if file.Name == exeName {
			return pkgInfo, file
		}
	}
	return nil, nil
}

func (ctrl *Controller) findExecFile(ctx context.Context, cfgFilePath, exeName string) (*config.Package, *config.PackageInfo, *config.File, error) {
	cfg := &config.Config{}
	if err := ctrl.ConfigReader.Read(cfgFilePath, cfg); err != nil {
		return nil, nil, nil, err //nolint:wrapcheck
	}
	if err := validate.Config(cfg); err != nil {
		return nil, nil, nil, fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.RegistryInstaller.InstallRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return nil, nil, nil, err //nolint:wrapcheck
	}
	for _, pkg := range cfg.Packages {
		if pkgInfo, file := ctrl.findExecFileFromPkg(registryContents, exeName, pkg); pkgInfo != nil {
			return pkg, pkgInfo, file, nil
		}
	}
	return nil, nil, nil, nil
}

func (ctrl *Controller) execCommand(ctx context.Context, exePath string, args []string) error {
	logE := ctrl.logE().WithField("exe_path", exePath)
	logE.Debug("execute the command")
	for i := 0; i < 10; i++ {
		logE.Debug("execute the command")
		cmd := exec.Command(exePath, args...)
		cmd.Stdin = ctrl.stdin
		cmd.Stdout = ctrl.stdout
		cmd.Stderr = ctrl.Stderr
		runner := timeout.NewRunner(0)
		if err := runner.Run(ctx, cmd); err != nil {
			exitCode := cmd.ProcessState.ExitCode()
			// https://pkg.go.dev/os#ProcessState.ExitCode
			// > ExitCode returns the exit code of the exited process,
			// > or -1 if the process hasn't exited or was terminated by a signal.
			if exitCode == -1 && ctx.Err() == nil {
				logE.WithField("retry_count", i+1).Debug("the process isn't started. retry")
				if err := wait(ctx, 10*time.Millisecond); err != nil { //nolint:gomnd
					return err
				}
				continue
			}
			logerr.WithError(logE, err).WithField("exit_code", exitCode).Debug("command was executed but it failed")
			return ecerror.Wrap(err, exitCode)
		}
		return nil
	}
	return nil
}
