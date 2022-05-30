package download

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/runtime"
	gh "github.com/google/go-github/v44/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type pkgDownloader struct {
	github  github.RepositoryService
	runtime *runtime.Runtime
	http    HTTPDownloader
}

func (downloader *pkgDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string, logE *logrus.Entry) (io.ReadCloser, error) {
	switch pkgInfo.GetType() {
	case config.PkgInfoTypeGitHubRelease:
		return downloader.getReadCloserFromGitHubRelease(ctx, pkg, pkgInfo, assetName, logE)
	case config.PkgInfoTypeGitHubContent:
		return downloader.getReadCloserFromGitHubContent(ctx, pkg, pkgInfo, assetName)
	case config.PkgInfoTypeGitHubArchive, config.PkgInfoTypeGo:
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
	if rc, err := downloader.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version)); err == nil {
		return rc, nil
	}
	// e.g. https://github.com/anqiansong/github-compare/archive/3972625c74bf6a5da00beb0e17e30e3e8d0c0950.zip
	if rc, err := downloader.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version)); err == nil {
		return rc, nil
	}
	u, _, err := downloader.github.GetArchiveLink(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, gh.Tarball, &gh.RepositoryContentGetOptions{
		Ref: pkg.Version,
	}, true)
	if err != nil {
		return nil, fmt.Errorf("git an archive link with GitHub API: %w", err)
	}
	return downloader.http.Download(ctx, u.String()) //nolint:wrapcheck
}

func (downloader *pkgDownloader) getReadCloserFromGitHubContent(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string) (io.ReadCloser, error) {
	return downloader.downloadGitHubContent(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
}

func (downloader *pkgDownloader) getReadCloserFromHTTP(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo) (io.ReadCloser, error) {
	uS, err := pkgInfo.RenderURL(pkg, downloader.runtime)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return downloader.http.Download(ctx, uS) //nolint:wrapcheck
}
