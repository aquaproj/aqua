package controller

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) download(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, dest, assetName string) error {
	logE := logrus.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	logE.Info("download and unarchive the package")

	if pkgInfo.Type == "github_release" && ctrl.GitHub == nil {
		return errGitHubTokenIsRequired
	}

	body, err := ctrl.downloadFromGitHub(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
	if err != nil {
		return err
	}
	return unarchive(body, assetName, pkgInfo.ArchiveType, dest)
}

var errGitHubTokenIsRequired = errors.New("GITHUB_TOKEN is required for the type `github_release`")
