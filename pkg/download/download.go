package download

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type RepositoriesService interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
	GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
}

func (downloader *PackageDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry) (io.ReadCloser, int64, error) {
	pkgInfo := pkg.PackageInfo
	switch pkgInfo.GetType() {
	case config.PkgInfoTypeGitHubRelease:
		pkgInfo := pkg.PackageInfo
		return downloader.downloadFromGitHubRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, assetName, logE)
	case config.PkgInfoTypeGitHubContent:
		pkgInfo := pkg.PackageInfo
		body, err := downloader.downloadGitHubContent(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, assetName)
		return body, 0, err
	case config.PkgInfoTypeGitHubArchive, config.PkgInfoTypeGo:
		return downloader.getReadCloserFromGitHubArchive(ctx, pkg)
	case config.PkgInfoTypeHTTP:
		uS, err := pkg.RenderURL(downloader.runtime)
		if err != nil {
			return nil, 0, err //nolint:wrapcheck
		}
		return downloader.http.Download(ctx, uS) //nolint:wrapcheck
	default:
		return nil, 0, logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}
}
