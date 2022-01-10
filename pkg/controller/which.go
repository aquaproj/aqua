package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) Which(ctx context.Context, param *Param, exeName string) error {
	which, err := ctrl.which(ctx, param, exeName)
	if err != nil {
		return err
	}
	fmt.Fprintln(ctrl.Stdout, which.ExePath)
	return nil
}

type Which struct {
	Package *Package
	PkgInfo *PackageInfo
	File    *File
	ExePath string
}

func (ctrl *Controller) which(ctx context.Context, param *Param, exeName string) (*Which, error) { //nolint:cyclop,funlen
	fields := logrus.Fields{
		"exe_name": exeName,
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get the current directory: %w", logerr.WithFields(err, fields))
	}

	for _, cfgFilePath := range ctrl.getConfigFilePaths(wd, param.ConfigFilePath) {
		pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
		if err != nil {
			return nil, err
		}
		if pkg != nil {
			return ctrl.whichFile(pkg, pkgInfo, file)
		}
	}

	for _, cfgFilePath := range getGlobalConfigFilePaths() {
		if _, err := os.Stat(cfgFilePath); err != nil {
			continue
		}
		pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
		if err != nil {
			return nil, err
		}
		if pkg != nil {
			return ctrl.whichFile(pkg, pkgInfo, file)
		}
	}

	cfgFilePath := ctrl.ConfigFinder.FindGlobal(ctrl.GlobalConfingDir)
	if _, err := os.Stat(cfgFilePath); err != nil {
		exePath := lookPath(exeName)
		if exePath == "" {
			return nil, logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
				"exe_name": exeName,
			})
		}
		return &Which{
			ExePath: exePath,
		}, nil
	}

	pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
	if err != nil {
		return nil, err
	}
	if pkg == nil {
		exePath := lookPath(exeName)
		if exePath == "" {
			return nil, logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
				"exe_name": exeName,
			})
		}
		return &Which{
			ExePath: exePath,
		}, nil
	}

	return ctrl.whichFile(pkg, pkgInfo, file)
}

func (ctrl *Controller) whichFile(pkg *Package, pkgInfo *PackageInfo, file *File) (*Which, error) {
	fileSrc, err := pkgInfo.GetFileSrc(pkg, file)
	if err != nil {
		return nil, fmt.Errorf("get file_src: %w", err)
	}
	pkgPath, err := pkgInfo.GetPkgPath(ctrl.RootDir, pkg)
	if err != nil {
		return nil, fmt.Errorf("get pkg install path: %w", err)
	}
	return &Which{
		Package: pkg,
		PkgInfo: pkgInfo,
		File:    file,
		ExePath: filepath.Join(pkgPath, fileSrc),
	}, nil
}
