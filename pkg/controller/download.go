package controller

import (
	"context"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type PackageDownloader interface {
	GetReadCloser(ctx context.Context, pkg *Package, pkgInfo PackageInfo, assetName string) (io.ReadCloser, error)
}

type pkgDownloader struct {
	GitHubRepositoryService GitHubRepositoryService
}

func (downloader *pkgDownloader) getReadCloserFromGitHubRelease(ctx context.Context, pkg *Package, pkgInfo PackageInfo, assetName string) (io.ReadCloser, error) {
	if downloader.GitHubRepositoryService == nil {
		return nil, errGitHubTokenIsRequired
	}
	p, ok := pkgInfo.(*GitHubReleasePackageInfo)
	if !ok {
		return nil, errGitHubReleaseTypeAssertion
	}
	return downloader.downloadFromGitHubRelease(ctx, p.RepoOwner, p.RepoName, pkg.Version, assetName)
}

func (downloader *pkgDownloader) getReadCloserFromGitHubContent(ctx context.Context, pkg *Package, pkgInfo PackageInfo, assetName string) (io.ReadCloser, error) {
	if downloader.GitHubRepositoryService == nil {
		return nil, errGitHubTokenIsRequired
	}
	p, ok := pkgInfo.(*GitHubContentPackageInfo)
	if !ok {
		return nil, errGitHubReleaseTypeAssertion
	}
	return downloader.downloadGitHubContent(ctx, p.RepoOwner, p.RepoName, pkg.Version, assetName)
}

func (downloader *pkgDownloader) getReadCloserFromHTTP(ctx context.Context, pkg *Package, pkgInfo PackageInfo) (io.ReadCloser, error) {
	p, ok := pkgInfo.(*HTTPPackageInfo)
	if !ok {
		return nil, errTypeAssertionHTTPPackageInfo
	}
	uS, err := p.RenderURL(pkg)
	if err != nil {
		return nil, err
	}
	return downloader.downloadFromURL(ctx, uS, http.DefaultClient)
}

func (downloader *pkgDownloader) GetReadCloser(ctx context.Context, pkg *Package, pkgInfo PackageInfo, assetName string) (io.ReadCloser, error) {
	switch pkgInfo.GetType() {
	case pkgInfoTypeGitHubRelease:
		return downloader.getReadCloserFromGitHubRelease(ctx, pkg, pkgInfo, assetName)
	case pkgInfoTypeGitHubContent:
		return downloader.getReadCloserFromGitHubContent(ctx, pkg, pkgInfo, assetName)
	case pkgInfoTypeHTTP:
		return downloader.getReadCloserFromHTTP(ctx, pkg, pkgInfo)
	default:
		return nil, logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}
}

func (ctrl *Controller) download(ctx context.Context, pkg *Package, pkgInfo PackageInfo, dest, assetName string) error {
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

	return unarchive(body, assetName, pkgInfo.GetFormat(), dest)
}
