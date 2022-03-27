package download

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type PackageDownloader interface {
	GetReadCloser(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string) (io.ReadCloser, error)
}

type PkgDownloader struct {
	GitHubRepositoryService github.RepositoryService
	logger                  *log.Logger
}

func (downloader *PkgDownloader) logE() *logrus.Entry {
	return downloader.logger.LogE()
}

func (downloader *PkgDownloader) getReadCloserFromGitHubRelease(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string) (io.ReadCloser, error) {
	return downloader.downloadFromGitHubRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
}

func (downloader *PkgDownloader) getReadCloserFromGitHubArchive(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo) (io.ReadCloser, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version)
	return FromURL(ctx, url, http.DefaultClient)
}

func (downloader *PkgDownloader) getReadCloserFromGitHubContent(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string) (io.ReadCloser, error) {
	return downloader.downloadGitHubContent(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
}

func (downloader *PkgDownloader) getReadCloserFromHTTP(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo) (io.ReadCloser, error) {
	uS, err := pkgInfo.RenderURL(pkg)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return FromURL(ctx, uS, http.DefaultClient)
}

func (downloader *PkgDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string) (io.ReadCloser, error) {
	switch pkgInfo.GetType() {
	case config.PkgInfoTypeGitHubRelease:
		return downloader.getReadCloserFromGitHubRelease(ctx, pkg, pkgInfo, assetName)
	case config.PkgInfoTypeGitHubContent:
		return downloader.getReadCloserFromGitHubContent(ctx, pkg, pkgInfo, assetName)
	case config.PkgInfoTypeGitHubArchive:
		return downloader.getReadCloserFromGitHubArchive(ctx, pkg, pkgInfo)
	case config.PkgInfoTypeHTTP:
		return downloader.getReadCloserFromHTTP(ctx, pkg, pkgInfo)
	default:
		return nil, logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}
}
