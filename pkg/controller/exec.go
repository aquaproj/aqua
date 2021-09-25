package controller

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func (ctrl *Controller) Exec(ctx context.Context, param *Param, args []string) error { //nolint:cyclop
	if len(args) == 0 {
		return errCommandIsRequired
	}

	exeName := filepath.Base(args[0])
	fields := logrus.Fields{
		"exe_name": exeName,
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", logerr.WithFields(err, fields))
	}

	if cfgFilePath := ctrl.getConfigFilePath(wd, param.ConfigFilePath); cfgFilePath != "" {
		pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
		if err != nil {
			return err
		}
		if pkg != nil {
			return ctrl.installAndExec(ctx, pkgInfo, pkg, file, args)
		}
	}

	for _, cfgFilePath := range getGlobalConfigFilePaths() {
		if _, err := os.Stat(cfgFilePath); err == nil {
			pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
			if err != nil {
				return err
			}
			if pkg != nil {
				return ctrl.installAndExec(ctx, pkgInfo, pkg, file, args)
			}
		}
	}

	cfgFilePath := ctrl.ConfigFinder.FindGlobal(ctrl.RootDir)
	if _, err := os.Stat(cfgFilePath); err != nil {
		return ctrl.findAndExecExtCommand(ctx, exeName, args[1:])
	}

	pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
	if err != nil {
		return err
	}
	if pkg == nil {
		return ctrl.findAndExecExtCommand(ctx, exeName, args[1:])
	}
	return ctrl.installAndExec(ctx, pkgInfo, pkg, file, args)
}

func (ctrl *Controller) findAndExecExtCommand(ctx context.Context, exeName string, args []string) error {
	exePath := lookPath(exeName)
	if exePath == "" {
		return logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
			"exe_name": exeName,
		})
	}
	return ctrl.execCommand(ctx, exePath, args)
}

func (ctrl *Controller) installAndExec(ctx context.Context, pkgInfo PackageInfo, pkg *Package, file *File, args []string) error {
	fileSrc, err := pkgInfo.GetFileSrc(pkg, file)
	if err != nil {
		return fmt.Errorf("get file_src: %w", err)
	}

	if err := ctrl.installPackage(ctx, pkgInfo, pkg, false); err != nil {
		return err
	}

	return ctrl.exec(ctx, pkg, pkgInfo, fileSrc, args[1:])
}

func (ctrl *Controller) findExecFileFromPkg(registries map[string]*RegistryContent, exeName string, pkg *Package) (PackageInfo, *File) {
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
		logE.WithError(err).Warnf("registry is invalid")
		return nil, nil
	}

	pkgInfo, ok := m[pkg.Name]
	if !ok {
		logE.Warn("package isn't found")
		return nil, nil
	}

	pkgInfo, err = pkgInfo.SetVersion(pkg.Version)
	if err != nil {
		logE.Warn("version constraint is invalid")
		return nil, nil
	}

	for _, file := range pkgInfo.GetFiles() {
		if file.Name == exeName {
			return pkgInfo, file
		}
	}
	return nil, nil
}

func (ctrl *Controller) findExecFile(ctx context.Context, cfgFilePath, exeName string) (*Package, PackageInfo, *File, error) {
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

func (ctrl *Controller) exec(ctx context.Context, pkg *Package, pkgInfo PackageInfo, src string, args []string) error {
	pkgPath, err := pkgInfo.GetPkgPath(ctrl.RootDir, pkg)
	if err != nil {
		return fmt.Errorf("get pkg install path: %w", err)
	}
	exePath := filepath.Join(pkgPath, src)

	if _, err := os.Stat(exePath); err != nil {
		return fmt.Errorf("file.src is invalid. file isn't found %s: %w", exePath, err)
	}

	return ctrl.execCommand(ctx, exePath, args)
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
		logE.WithError(err).WithField("exit_code", exitCode).Debug("command was executed but it failed")
		return ecerror.Wrap(err, exitCode)
	}
	return nil
}
