package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type File struct {
	Type      string
	RepoOwner string
	RepoName  string
	Version   string
	Asset     string
	URL       string
	Path      string
	Private   bool
}

type Downloader struct {
	github    GitHub
	http      HTTPDownloader
	ghContent domain.GitHubContentFileDownloader
	ghRelease domain.GitHubReleaseDownloader
}

type GitHub interface {
	DownloadContents(ctx context.Context, logE *logrus.Entry, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
	GetReleaseByTag(ctx context.Context, logE *logrus.Entry, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, logE *logrus.Entry, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
	GetArchiveLink(ctx context.Context, logE *logrus.Entry, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, maxRedirects int) (*url.URL, *github.Response, error)
}

func NewDownloader(gh GitHub, httpDownloader HTTPDownloader) *Downloader {
	return &Downloader{
		github:    gh,
		http:      httpDownloader,
		ghContent: NewGitHubContentFileDownloader(gh, httpDownloader),
		ghRelease: NewGitHubReleaseDownloader(gh, httpDownloader),
	}
}

type ClientAPI interface {
	ReadCloser(ctx context.Context, logE *logrus.Entry, file *File) (io.ReadCloser, int64, error)
}

func (dl *Downloader) ReadCloser(ctx context.Context, logE *logrus.Entry, file *File) (io.ReadCloser, int64, error) {
	switch file.Type {
	case config.PkgInfoTypeGitHubRelease:
		return dl.ghRelease.DownloadGitHubRelease(ctx, logE, &domain.DownloadGitHubReleaseParam{ //nolint:wrapcheck
			RepoOwner: file.RepoOwner,
			RepoName:  file.RepoName,
			Version:   file.Version,
			Asset:     file.Asset,
			Private:   file.Private,
		})
	case config.PkgInfoTypeGitHubContent:
		file, err := dl.ghContent.DownloadGitHubContentFile(ctx, logE, &domain.GitHubContentFileParam{
			RepoOwner: file.RepoOwner,
			RepoName:  file.RepoName,
			Ref:       file.Version,
			Path:      file.Path,
			Private:   file.Private,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("download a package from GitHub Content: %w", err)
		}
		if file.ReadCloser != nil {
			return file.ReadCloser, 0, nil
		}
		return io.NopCloser(strings.NewReader(file.String)), 0, nil
	case config.PkgInfoTypeGitHubArchive:
		return dl.getReadCloserFromGitHubArchive(ctx, logE, file)
	case config.PkgInfoTypeHTTP:
		rc, code, err := dl.http.Download(ctx, file.URL)
		if err != nil {
			return rc, code, fmt.Errorf("download a package: %w", logerr.WithFields(err, logrus.Fields{
				"download_url": file.URL,
			}))
		}
		return rc, code, nil
	default:
		return nil, 0, logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"file_type": file.Type,
		})
	}
}
