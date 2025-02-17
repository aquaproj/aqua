package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var errUnknownChecksumFileType = errors.New("unknown checksum type")

type ChecksumDownloaderImpl struct {
	github    GitHub
	runtime   *runtime.Runtime
	http      HTTPDownloader
	ghRelease domain.GitHubReleaseDownloader
}

type GitHub interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, maxRedirects int) (*url.URL, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
}

func NewChecksumDownloader(gh GitHub, rt *runtime.Runtime, httpDownloader HTTPDownloader) *ChecksumDownloaderImpl {
	return &ChecksumDownloaderImpl{
		github:    gh,
		runtime:   rt,
		http:      httpDownloader,
		ghRelease: NewGitHubReleaseDownloader(gh, httpDownloader),
	}
}

type ChecksumDownloader interface {
	DownloadChecksum(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, pkg *config.Package) (io.ReadCloser, int64, error)
}

func (dl *ChecksumDownloaderImpl) DownloadChecksum(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, pkg *config.Package) (io.ReadCloser, int64, error) {
	pkgInfo := pkg.PackageInfo
	switch pkg.PackageInfo.Checksum.Type {
	case config.PkgInfoTypeGitHubRelease:
		asset, err := pkg.RenderChecksumFileName(rt)
		if err != nil {
			return nil, 0, fmt.Errorf("render a checksum file name: %w", err)
		}
		return dl.ghRelease.DownloadGitHubRelease(ctx, logE, &domain.DownloadGitHubReleaseParam{ //nolint:wrapcheck
			RepoOwner: pkgInfo.RepoOwner,
			RepoName:  pkgInfo.RepoName,
			Version:   pkg.Package.Version,
			Asset:     asset,
		})
	case config.PkgInfoTypeHTTP:
		u, err := pkg.RenderChecksumURL(rt)
		if err != nil {
			return nil, 0, fmt.Errorf("render a checksum file name: %w", err)
		}
		rc, code, err := dl.http.Download(ctx, u)
		if err != nil {
			return rc, code, fmt.Errorf("download a checksum file: %w", logerr.WithFields(err, logrus.Fields{
				"download_url": u,
			}))
		}
		return rc, code, nil
	default:
		return nil, 0, logerr.WithFields(errUnknownChecksumFileType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.Type,
		})
	}
}
