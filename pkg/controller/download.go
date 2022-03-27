package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type PackageDownloader interface {
	GetReadCloser(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, assetName string) (io.ReadCloser, error)
}

type pkgDownloader struct {
	GitHubRepositoryService GitHubRepositoryService
	logE                    func() *logrus.Entry
}

func (downloader *pkgDownloader) getReadCloserFromGitHubRelease(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, assetName string) (io.ReadCloser, error) {
	return downloader.downloadFromGitHubRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
}

func (downloader *pkgDownloader) getReadCloserFromGitHubArchive(ctx context.Context, pkg *Package, pkgInfo *PackageInfo) (io.ReadCloser, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version)
	return downloadFromURL(ctx, url, http.DefaultClient)
}

func (downloader *pkgDownloader) getReadCloserFromGitHubContent(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, assetName string) (io.ReadCloser, error) {
	return downloader.downloadGitHubContent(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
}

func (downloader *pkgDownloader) getReadCloserFromHTTP(ctx context.Context, pkg *Package, pkgInfo *PackageInfo) (io.ReadCloser, error) {
	uS, err := pkgInfo.renderURL(pkg)
	if err != nil {
		return nil, err
	}
	return downloadFromURL(ctx, uS, http.DefaultClient)
}

func (downloader *pkgDownloader) GetReadCloser(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, assetName string) (io.ReadCloser, error) {
	switch pkgInfo.GetType() {
	case pkgInfoTypeGitHubRelease:
		return downloader.getReadCloserFromGitHubRelease(ctx, pkg, pkgInfo, assetName)
	case pkgInfoTypeGitHubContent:
		return downloader.getReadCloserFromGitHubContent(ctx, pkg, pkgInfo, assetName)
	case pkgInfoTypeGitHubArchive:
		return downloader.getReadCloserFromGitHubArchive(ctx, pkg, pkgInfo)
	case pkgInfoTypeHTTP:
		return downloader.getReadCloserFromHTTP(ctx, pkg, pkgInfo)
	default:
		return nil, logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}
}

func (ctrl *Controller) download(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, dest, assetName string) error {
	logE := ctrl.logE().WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	logE.Info("download and unarchive the package")

	body, err := ctrl.PackageDownloader.GetReadCloser(ctx, pkg, pkgInfo, assetName)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err //nolint:wrapcheck
	}

	return unarchive.Unarchive(body, assetName, pkgInfo.GetFormat(), dest) //nolint:wrapcheck
}
