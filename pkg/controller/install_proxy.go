package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) installProxy(ctx context.Context) error {
	pkg := &Package{
		Name:     "aqua-proxy",
		Version:  "v0.1.0-1",
		Registry: "inline",
	}
	logE := logrus.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})

	logE.Debug("install the proxy")
	pkgInfo := &PackageInfo{
		Name:      "inline",
		Type:      "github_release",
		RepoOwner: "suzuki-shunsuke",
		RepoName:  "aqua-proxy",
		Asset:     nil,
		Files: []*File{
			{
				Name: "aqua-proxy",
			},
		},
	}

	assetName := "aqua-proxy_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"

	pkgPath := getPkgPath(ctrl.RootDir, pkg, pkgInfo, assetName)
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
	if err := os.Symlink(filepath.Join(pkgPath, "aqua-proxy"), filepath.Join(ctrl.RootDir, "bin", "aqua-proxy")); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}

	return nil
}
