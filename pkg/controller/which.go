package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) Which(ctx context.Context, param *Param, args []string) error { //nolint:cyclop,funlen
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
			return ctrl.which(pkg, pkgInfo, file)
		}
	}

	for _, cfgFilePath := range getGlobalConfigFilePaths() {
		if _, err := os.Stat(cfgFilePath); err != nil {
			continue
		}
		pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
		if err != nil {
			return err
		}
		if pkg != nil {
			return ctrl.which(pkg, pkgInfo, file)
		}
	}

	cfgFilePath := ctrl.ConfigFinder.FindGlobal(ctrl.RootDir)
	if _, err := os.Stat(cfgFilePath); err != nil {
		exePath := lookPath(exeName)
		if exePath == "" {
			return logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
				"exe_name": exeName,
			})
		}
		fmt.Fprintln(ctrl.Stdout, exePath)
		return nil
	}

	pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
	if err != nil {
		return err
	}
	if pkg == nil {
		exePath := lookPath(exeName)
		if exePath == "" {
			return logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
				"exe_name": exeName,
			})
		}
		fmt.Fprintln(ctrl.Stdout, exePath)
		return nil
	}

	return ctrl.which(pkg, pkgInfo, file)
}

func (ctrl *Controller) which(pkg *Package, pkgInfo *MergedPackageInfo, file *File) error {
	fileSrc, err := pkgInfo.GetFileSrc(pkg, file)
	if err != nil {
		return fmt.Errorf("get file_src: %w", err)
	}
	pkgPath, err := pkgInfo.GetPkgPath(ctrl.RootDir, pkg)
	if err != nil {
		return fmt.Errorf("get pkg install path: %w", err)
	}
	fmt.Fprintln(ctrl.Stdout, filepath.Join(pkgPath, fileSrc))
	return nil
}
