package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/aqua/pkg/log"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var (
	errCommandIsRequired = errors.New("command is required")
	errCommandIsNotFound = errors.New("command is not found")
)

func (ctrl *Controller) Exec(ctx context.Context, param *Param, args []string) error { //nolint:funlen,cyclop,gocognit
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

	cfgFilePath := ctrl.getConfigFilePath(wd, param.ConfigFilePath)

	binDir := filepath.Join(ctrl.RootDir, "bin")

	var (
		pkg            *Package
		pkgInfo        PackageInfo
		file           *File
		inlineRegistry map[string]PackageInfo
	)
	if cfgFilePath != "" { //nolint:nestif
		cfg := &Config{}
		if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
			return err
		}

		inlineRegistry = make(map[string]PackageInfo, len(cfg.InlineRegistry))
		for _, pkgInfo := range cfg.InlineRegistry {
			inlineRegistry[pkgInfo.GetName()] = pkgInfo
		}

		pkg, pkgInfo, file = ctrl.findExecFile(inlineRegistry, cfg, exeName)
		if pkg == nil {
			cfgFilePath = ctrl.ConfigFinder.FindGlobal(ctrl.RootDir)
			if _, err := os.Stat(cfgFilePath); err != nil {
				exePath := lookPath(exeName)
				if exePath == "" {
					return errCommandIsNotFound
				}
				return ctrl.execCommand(ctx, exePath, args[1:])
			}
			cfg := &Config{}
			if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
				return err
			}

			inlineRegistry = make(map[string]PackageInfo, len(cfg.InlineRegistry))
			for _, pkgInfo := range cfg.InlineRegistry {
				inlineRegistry[pkgInfo.GetName()] = pkgInfo
			}

			pkg, pkgInfo, file = ctrl.findExecFile(inlineRegistry, cfg, exeName)
			if pkg == nil {
				exePath := lookPath(exeName)
				if exePath == "" {
					return errCommandIsNotFound
				}
				return ctrl.execCommand(ctx, exePath, args[1:])
			}
		}
	} else {
		cfgFilePath = ctrl.ConfigFinder.FindGlobal(ctrl.RootDir)
		if _, err := os.Stat(cfgFilePath); err != nil {
			exePath := lookPath(exeName)
			if exePath == "" {
				return logerr.WithFields(errCommandIsNotFound, fields) //nolint:wrapcheck
			}
			return ctrl.execCommand(ctx, exePath, args[1:])
		}
		cfg := &Config{}
		if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
			return err
		}

		inlineRegistry = make(map[string]PackageInfo, len(cfg.InlineRegistry))
		for _, pkgInfo := range cfg.InlineRegistry {
			inlineRegistry[pkgInfo.GetName()] = pkgInfo
		}

		pkg, pkgInfo, file = ctrl.findExecFile(inlineRegistry, cfg, exeName)
		if pkg == nil {
			exePath := lookPath(exeName)
			if exePath == "" {
				return logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
					"exe_name": exeName,
				})
			}
			return ctrl.execCommand(ctx, exePath, args[1:])
		}
	}

	fileSrc, err := pkgInfo.GetFileSrc(pkg, file)
	if err != nil {
		return fmt.Errorf("get file_src: %w", err)
	}

	if err := ctrl.installPackage(ctx, inlineRegistry, pkg, binDir, false, false); err != nil {
		return err
	}

	return ctrl.exec(ctx, pkg, pkgInfo, fileSrc, args[1:])
}

func (ctrl *Controller) findExecFile(inlineRegistry map[string]PackageInfo, cfg *Config, exeName string) (*Package, PackageInfo, *File) {
	for _, pkg := range cfg.Packages {
		pkgInfo, ok := inlineRegistry[pkg.Name]
		if !ok {
			log.New().Warnf("registry isn't found %s", pkg.Name)
			continue
		}
		for _, file := range pkgInfo.GetFiles() {
			if file.Name == exeName {
				return pkg, pkgInfo, file
			}
		}
	}
	return nil, nil, nil
}

func isUnarchived(archiveType, assetName string) bool {
	return archiveType == "raw" || (archiveType == "" && filepath.Ext(assetName) == "")
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

	logE := log.New().WithField("exe_path", exePath)
	logE.Debug("execute the command")
	if err := runner.Run(ctx, cmd); err != nil {
		exitCode := cmd.ProcessState.ExitCode()
		logE.WithError(err).WithField("exit_code", exitCode).Debug("command was executed but it failed")
		return ecerror.Wrap(err, exitCode)
	}
	return nil
}
