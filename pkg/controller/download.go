package controller

import (
	"context"
	"errors"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/aqua/pkg/log"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) download(ctx context.Context, pkg *Package, pkgInfo PackageInfo, dest, assetName string) error {
	logE := log.New().WithFields(logrus.Fields{
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
			return errors.New("pkg typs is github_release, but it isn't *GitHubReleasePackageInfo")
		}
		b, err := ctrl.downloadFromGitHub(ctx, p.RepoOwner, p.RepoName, pkg.Version, assetName)
		if err != nil {
			return err
		}
		body = b
	case pkgInfoTypeHTTP:
		p, ok := pkgInfo.(*HTTPPackageInfo)
		if !ok {
			return errors.New("pkg typs is http, but it isn't *HTTPPackageInfo")
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
		return logerr.WithFields(errors.New("invalid type"), logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}

	defer body.Close()
	return unarchive(body, assetName, pkgInfo.GetArchiveType(), dest)
}

var errGitHubTokenIsRequired = errors.New("GITHUB_TOKEN is required for the type `github_release`")
