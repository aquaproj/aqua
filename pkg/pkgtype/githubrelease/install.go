package githubrelease

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/render"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *Installer) Install(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, logE *logrus.Entry) error {
	if err := inst.validate(pkgInfo); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	assetName, err := inst.assetName(pkgInfo, pkg)
	if err != nil {
		return err
	}

	body, err := inst.githubRelease.Download(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName, logE)
	if err != nil {
		return fmt.Errorf("download an asset from GitHub Releases: %w", logerr.WithFields(err, logrus.Fields{
			"asset_name": assetName,
		}))
	}
	defer body.Close()

	dest := inst.getInstallDir(pkg, pkgInfo, assetName)

	if err := unarchive.Unarchive(&unarchive.File{
		Body:     body,
		Filename: assetName,
		Type:     pkgInfo.Format,
	}, dest, logE, inst.fs); err != nil {
		return fmt.Errorf("unarchive a file: %w", logerr.WithFields(err, logrus.Fields{
			"file_format": pkgInfo.Format,
			"asset_name":  assetName,
		}))
	}
	return nil
}

func (inst *Installer) CheckInstalled(pkg *config.Package, pkgInfo *config.PackageInfo) (bool, error) {
	assetName, err := inst.assetName(pkgInfo, pkg)
	if err != nil {
		return false, err
	}
	f, err := util.ExistDir(inst.fs, inst.getInstallDir(pkg, pkgInfo, assetName))
	if err != nil {
		return false, fmt.Errorf("check if the directory exists: %w", err)
	}
	return f, nil
}

func (inst *Installer) getInstallDir(pkg *config.Package, pkgInfo *config.PackageInfo, assetName string) string {
	return filepath.Join(inst.rootDir, "pkgs", PkgType, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
}

func (inst *Installer) assetName(pkgInfo *config.PackageInfo, pkg *config.Package) (string, error) {
	return template.Execute(*pkgInfo.Asset, map[string]interface{}{ //nolint:wrapcheck
		"Version": pkg.Version,
		"GOOS":    inst.runtime.GOOS,
		"GOARCH":  inst.runtime.GOARCH,
		"OS":      render.Replace(inst.runtime.GOOS, pkgInfo.Replacements),
		"Arch":    render.GetArch(pkgInfo.GetRosetta2(), pkgInfo.Replacements, inst.runtime),
		"Format":  pkgInfo.GetFormat(),
	})
}

func (inst *Installer) GetFiles(pkgInfo *config.PackageInfo) []*config.File {
	return pkgInfo.GetFiles()
}
