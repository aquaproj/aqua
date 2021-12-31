package controller

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func getGlobalConfigFilePaths() []string {
	src := strings.Split(os.Getenv("AQUA_GLOBAL_CONFIG"), ":")
	paths := make([]string, 0, len(src))
	for _, s := range src {
		if s == "" {
			continue
		}
		paths = append(paths, s)
	}
	return paths
}

func (ctrl *Controller) Exec(ctx context.Context, param *Param, exeName string, args []string) error {
	which, err := ctrl.which(ctx, param, exeName)
	if err != nil {
		return err
	}
	if which.Package != nil {
		if err := ctrl.installPackage(ctx, which.PkgInfo, which.Package, false); err != nil {
			return err
		}
	}
	return ctrl.execCommand(ctx, which.ExePath, args)
}

func (ctrl *Controller) findExecFileFromPkg(registries map[string]*RegistryContent, exeName string, pkg *Package) (*MergedPackageInfo, *File) {
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
		supported, err := pkgInfo.SupportedIf.Check(pkg.Version)
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

func (ctrl *Controller) findExecFile(ctx context.Context, cfgFilePath, exeName string) (*Package, *MergedPackageInfo, *File, error) {
	cfg := &Config{}
	if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
		return nil, nil, nil, err
	}
	if err := validateConfig(cfg); err != nil {
		return nil, nil, nil, fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.installRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, pkg := range cfg.Packages {
		if pkgInfo, file := ctrl.findExecFileFromPkg(registryContents, exeName, pkg); pkgInfo != nil {
			return pkg, pkgInfo, file, nil
		}
	}
	return nil, nil, nil, nil
}

func (ctrl *Controller) execCommand(ctx context.Context, exePath string, args []string) error {
	cmd := exec.Command(exePath, args...)
	cmd.Stdin = ctrl.Stdin
	cmd.Stdout = ctrl.Stdout
	cmd.Stderr = ctrl.Stderr
	runner := timeout.NewRunner(0)

	logE := ctrl.logE().WithField("exe_path", exePath)
	logE.Debug("execute the command")
	if err := runner.Run(ctx, cmd); err != nil {
		exitCode := cmd.ProcessState.ExitCode()
		logerr.WithError(logE, err).WithField("exit_code", exitCode).Debug("command was executed but it failed")
		return ecerror.Wrap(err, exitCode)
	}
	return nil
}
