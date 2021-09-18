package controller

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) download(ctx context.Context, pkg *Package, pkgInfo PackageInfo, dest, assetName string) error {
	logE := ctrl.logE().WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	logE.Info("download and unarchive the package")

	var body io.ReadCloser
	switch pkgInfo.GetType() {
	case pkgInfoTypeGitHubRelease:
		if ctrl.GitHub == nil {
			return errGitHubTokenIsRequired
		}
		p, ok := pkgInfo.(*GitHubReleasePackageInfo)
		if !ok {
			return errGitHubReleaseTypeAssertion
		}
		b, err := ctrl.downloadFromGitHub(ctx, p.RepoOwner, p.RepoName, pkg.Version, assetName)
		if err != nil {
			return err
		}
		body = b
	case pkgInfoTypeHTTP:
		p, ok := pkgInfo.(*HTTPPackageInfo)
		if !ok {
			return errTypeAssertionHTTPPackageInfo
		}
		uS, err := p.RenderURL(pkg)
		if err != nil {
			return err
		}
		b, err := ctrl.downloadFromURL(ctx, uS)
		if err != nil {
			return err
		}
		body = b
	default:
		return logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}

	defer body.Close()
	return unarchive(body, assetName, pkgInfo.GetFormat(), dest)
}
