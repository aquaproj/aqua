package download

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type pkgDownloader struct {
	github  github.RepositoryService
	runtime *runtime.Runtime
}

func (downloader *pkgDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string, logE *logrus.Entry) (io.ReadCloser, error) {
	switch pkgInfo.GetType() {
	case config.PkgInfoTypeGitHubRelease:
		return downloader.getReadCloserFromGitHubRelease(ctx, pkg, pkgInfo, assetName, logE)
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

func (downloader *pkgDownloader) getReadCloserFromGitHubRelease(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string, logE *logrus.Entry) (io.ReadCloser, error) {
	return downloader.downloadFromGitHubRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName, logE)
}

func (downloader *pkgDownloader) getReadCloserFromGitHubArchive(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo) (io.ReadCloser, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version)
	return fromURL(ctx, url, http.DefaultClient)
}

func (downloader *pkgDownloader) getReadCloserFromGitHubContent(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string) (io.ReadCloser, error) {
	return downloader.downloadGitHubContent(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
}

func (downloader *pkgDownloader) getReadCloserFromHTTP(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo) (io.ReadCloser, error) {
	uS, err := pkgInfo.RenderURL(pkg, downloader.runtime)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return fromURL(ctx, uS, http.DefaultClient)
}
