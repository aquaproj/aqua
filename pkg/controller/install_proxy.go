package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/aqua/pkg/log"
	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
)

func (ctrl *Controller) installProxy(ctx context.Context) error {
	pkg := &Package{
		Name:     proxyName,
		Version:  "v0.1.2", // renovate: depName=suzuki-shunsuke/aqua-proxy
		Registry: "inline",
	}
	logE := log.New().WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})

	assetNameTpl, err := text.New(`aqua-proxy_{{.OS}}_{{.Arch}}.tar.gz`)
	if err != nil {
		return fmt.Errorf("render the asset name of aqua-proxy: %w", err)
	}

	logE.Debug("install the proxy")
	pkgInfo := &GitHubReleasePackageInfo{
		Name:      "inline",
		RepoOwner: "suzuki-shunsuke",
		RepoName:  proxyName,
		Asset:     assetNameTpl,
		Files: []*File{
			{
				Name: proxyName,
			},
		},
	}
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return err
	}

	pkgPath, err := pkgInfo.GetPkgPath(ctrl.RootDir, pkg)
	if err != nil {
		return err
	}
	logE.Debug("check if the package is already installed")
	finfo, err := os.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		if err := ctrl.download(ctx, pkg, pkgInfo, pkgPath, assetName); err != nil {
			return err
		}
	} else {
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	// create a symbolic link
	a, err := filepath.Rel(filepath.Join(ctrl.RootDir, "bin"), filepath.Join(pkgPath, proxyName))
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}
	if err := os.Symlink(a, filepath.Join(ctrl.RootDir, "bin", proxyName)); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}

	return nil
}
