package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type pkgDownloader struct {
	github  RepositoriesService
	runtime *runtime.Runtime
	http    HTTPDownloader
}

type RepositoriesService interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
	GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
}

func (downloader *pkgDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry) (io.ReadCloser, int64, error) {
	pkgInfo := pkg.PackageInfo
	switch pkgInfo.GetType() {
	case config.PkgInfoTypeGitHubRelease:
		return downloader.getReadCloserFromGitHubRelease(ctx, pkg, assetName, logE)
	case config.PkgInfoTypeGitHubContent:
		body, err := downloader.getReadCloserFromGitHubContent(ctx, pkg, assetName)
		return body, 0, err
	case config.PkgInfoTypeGitHubArchive, config.PkgInfoTypeGo:
		return downloader.getReadCloserFromGitHubArchive(ctx, pkg)
	case config.PkgInfoTypeHTTP:
		return downloader.getReadCloserFromHTTP(ctx, pkg)
	default:
		return nil, 0, logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}
}

func (downloader *pkgDownloader) getReadCloserFromGitHubRelease(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry) (io.ReadCloser, int64, error) {
	pkgInfo := pkg.PackageInfo
	return downloader.downloadFromGitHubRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, assetName, logE)
}

func (downloader *pkgDownloader) getReadCloserFromGitHubArchive(ctx context.Context, pkg *config.Package) (io.ReadCloser, int64, error) {
	pkgInfo := pkg.PackageInfo
	if rc, length, err := downloader.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version)); err == nil {
		return rc, length, nil
	}
	// e.g. https://github.com/anqiansong/github-compare/archive/3972625c74bf6a5da00beb0e17e30e3e8d0c0950.zip
	if rc, length, err := downloader.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version)); err == nil {
		return rc, length, nil
	}
	u, _, err := downloader.github.GetArchiveLink(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, github.Tarball, &github.RepositoryContentGetOptions{
		Ref: pkg.Package.Version,
	}, true)
	if err != nil {
		return nil, 0, fmt.Errorf("git an archive link with GitHub API: %w", err)
	}
	return downloader.http.Download(ctx, u.String()) //nolint:wrapcheck
}

func (downloader *pkgDownloader) getReadCloserFromGitHubContent(ctx context.Context, pkg *config.Package, assetName string) (io.ReadCloser, error) {
	pkgInfo := pkg.PackageInfo
	body, err := downloader.downloadGitHubContent(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, assetName)
	return body, err
}

func (downloader *pkgDownloader) getReadCloserFromHTTP(ctx context.Context, pkg *config.Package) (io.ReadCloser, int64, error) {
	uS, err := pkg.RenderURL(downloader.runtime)
	if err != nil {
		return nil, 0, err //nolint:wrapcheck
	}
	return downloader.http.Download(ctx, uS) //nolint:wrapcheck
}
