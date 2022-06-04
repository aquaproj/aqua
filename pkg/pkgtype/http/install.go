package http

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *Installer) Install(ctx context.Context, pkg *config.Package, logE *logrus.Entry) error {
	pkgInfo := pkg.PackageInfo
	if err := inst.validate(pkgInfo); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	uS, err := inst.renderURL(pkg)
	if err != nil {
		return fmt.Errorf("render URL: %w", err)
	}

	body, err := inst.http.Download(ctx, uS)
	if err != nil {
		return fmt.Errorf("download a file: %w", err)
	}
	defer body.Close()

	u, err := url.Parse(uS)
	if err != nil {
		return fmt.Errorf("parse the URL: %w", err)
	}

	dest := inst.getInstallDir(u)

	fileName := filepath.Base(u.Path)
	if err := unarchive.Unarchive(&unarchive.File{
		Body:     body,
		Filename: fileName,
		Type:     pkgInfo.Format,
	}, dest, logE, inst.fs); err != nil {
		return fmt.Errorf("unarchive a file: %w", logerr.WithFields(err, logrus.Fields{
			"file_name":   fileName,
			"file_format": pkgInfo.Format,
		}))
	}
	return nil
}

func (inst *Installer) getInstallDir(u *url.URL) string {
	return filepath.Join(inst.rootDir, "pkgs", PkgType, u.Host, u.Path)
}

func (inst *Installer) getURL(pkg *config.Package) (*url.URL, error) {
	uS, err := inst.renderURL(pkg)
	if err != nil {
		return nil, fmt.Errorf("render URL: %w", err)
	}
	u, err := url.Parse(uS)
	if err != nil {
		return nil, fmt.Errorf("parse the URL: %w", err)
	}
	return u, nil
}

func (inst *Installer) CheckInstalled(pkg *config.Package) (bool, error) {
	u, err := inst.getURL(pkg)
	if err != nil {
		return false, err
	}

	return util.ExistDir(inst.fs, inst.getInstallDir(u)) //nolint:wrapcheck
}

func (inst *Installer) GetFiles(pkgInfo *registry.PackageInfo) []*registry.File {
	if files := pkgInfo.GetFiles(); len(files) != 0 {
		return files
	}
	return nil
}
