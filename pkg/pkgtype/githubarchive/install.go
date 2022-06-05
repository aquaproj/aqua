package githubarchive

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
	if err := inst.validate(pkg.PackageInfo); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	body, err := inst.githubArchive.Download(ctx, pkg)
	if err != nil {
		return fmt.Errorf("download a GitHub Repository Archive: %w", err)
	}
	defer body.Close()

	dest := inst.getInstallDir(pkg)

	if err := unarchive.Unarchive(&unarchive.File{
		Body: body,
		Type: "tar.gz",
	}, dest, logE, inst.fs); err != nil {
		return fmt.Errorf("unarchive a tarball: %w", err)
	}
	return nil
}

func (inst *Installer) CheckInstalled(pkg *config.Package) (bool, error) {
	return util.ExistDir(inst.fs, inst.getInstallDir(pkg)) //nolint:wrapcheck
}

func (inst *Installer) getInstallDir(pkg *config.Package) string {
	pkgInfo := pkg.PackageInfo
	return filepath.Join(inst.rootDir, "pkgs", PkgType, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version)
}

func (inst *Installer) GetFiles(pkgInfo *registry.PackageInfo) []*registry.File {
	return pkgInfo.GetFiles()
}
