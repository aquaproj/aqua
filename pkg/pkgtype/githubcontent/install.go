package githubcontent

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) Install(ctx context.Context, pkg *config.Package, logE *logrus.Entry) error {
	pkgInfo := pkg.PackageInfo
	if err := inst.validate(pkgInfo); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	filePath := *pkgInfo.Path

	body, err := inst.githubContent.Download(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, filePath, logE)
	if err != nil {
		return fmt.Errorf("download a file from GitHub Repository: %w", err)
	}
	defer body.Close()

	dest := filepath.Join(inst.getInstallDir(pkg), filePath)

	if err := unarchive.Unarchive(&unarchive.File{
		Body:     body,
		Filename: filepath.Base(filePath),
		Type:     pkgInfo.Format,
	}, dest, logE, inst.fs); err != nil {
		return fmt.Errorf("unarchive a file: %w", err)
	}
	return nil
}

func (inst *Installer) GetFiles(pkgInfo *registry.PackageInfo) []*registry.File {
	return pkgInfo.GetFiles()
}

func (inst *Installer) CheckInstalled(pkg *config.Package) (bool, error) {
	f, err := util.ExistDir(inst.fs, inst.getInstallDir(pkg))
	if err != nil {
		return false, fmt.Errorf("check the directory exists: %w", err)
	}
	return f, nil
}

func (inst *Installer) getInstallDir(pkg *config.Package) string {
	pkgInfo := pkg.PackageInfo
	return filepath.Join(inst.rootDir, "pkgs", PkgType, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, *pkgInfo.Path)
}
