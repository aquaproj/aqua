package controller

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) download(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, dest, assetName string) error {
	logE := logrus.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"repository":      pkg.Repository,
	})
	logE.Info("download and unarchive the package")
	s := strings.Split(pkgInfo.Repo, "/")
	owner := s[0]
	repoName := s[1]

	body, err := ctrl.downloadFromGitHub(ctx, owner, repoName, pkg.Version, assetName)
	if err != nil {
		return err
	}
	if err := unarchive(body, assetName, pkgInfo.ArchiveType, dest); err != nil {
		return err
	}
	return nil
}
