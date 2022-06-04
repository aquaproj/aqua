package goinstall

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) Install(ctx context.Context, pkg *config.Package, logE *logrus.Entry) error {
	pkgInfo := pkg.PackageInfo
	if err := inst.validate(pkg); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	goPkgPath := inst.getGoPath(pkgInfo) + "@" + pkg.Package.Version

	binDir := inst.getBinDir(pkg)
	binNames := inst.getBinNames(pkgInfo)

	if inst.exist(binDir, binNames) {
		return nil
	}

	logE.WithFields(logrus.Fields{
		"gobin":           binDir,
		"go_package_path": goPkgPath,
	}).Info("Installing a Go tool")
	if _, err := inst.installer.GoInstall(ctx, goPkgPath, binDir); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}

func (inst *Installer) CheckInstalled(pkg *config.Package) (bool, error) {
	pkgInfo := pkg.PackageInfo
	for _, file := range inst.GetFiles(pkgInfo) {
		filePath, err := inst.GetFilePath(pkg, file)
		if err != nil {
			return false, err
		}
		f, err := util.ExistFile(inst.fs, filePath)
		if err != nil {
			return false, err //nolint:wrapcheck
		}
		if !f {
			return false, nil
		}
	}
	return true, nil
}

func (inst *Installer) getPkgPath(pkgInfo *registry.PackageInfo) string {
	if pkgInfo.Path != nil {
		return *pkgInfo.Path
	}
	if pkgInfo.HasRepo() {
		return "github.com/" + pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
	}
	return ""
}

func (inst *Installer) GetFiles(pkgInfo *registry.PackageInfo) []*registry.File {
	if files := pkgInfo.GetFiles(); len(files) != 0 {
		return files
	}
	if pkgInfo.Asset != nil {
		return []*registry.File{
			{
				Name: *pkgInfo.Asset,
			},
		}
	}
	return []*registry.File{
		{
			Name: filepath.Base(inst.getPkgPath(pkgInfo)),
		},
	}
}
