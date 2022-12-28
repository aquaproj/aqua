package download

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
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
	github    domain.RepositoriesService
	http      HTTPDownloader
	ghContent domain.GitHubContentFileDownloader
	ghRelease domain.GitHubReleaseDownloader
}

func NewDownloader(gh domain.RepositoriesService, httpDownloader HTTPDownloader) *Downloader {
	return &Downloader{
		github:    gh,
		http:      httpDownloader,
		ghContent: NewGitHubContentFileDownloader(gh, httpDownloader),
		ghRelease: NewGitHubReleaseDownloader(gh, httpDownloader),
	}
}

type ClientAPI interface {
	GetReadCloser(ctx context.Context, logE *logrus.Entry, file *File) (io.ReadCloser, int64, error)
}

type Mock struct {
	RC   io.ReadCloser
	Code int64
	Err  error
}

func (mock *Mock) GetReadCloser(ctx context.Context, logE *logrus.Entry, file *File) (io.ReadCloser, int64, error) {
	return mock.RC, mock.Code, mock.Err
}

func (downloader *Downloader) GetReadCloser(ctx context.Context, logE *logrus.Entry, file *File) (io.ReadCloser, int64, error) {
	switch file.Type {
	case config.PkgInfoTypeGitHubRelease:
		return downloader.ghRelease.DownloadGitHubRelease(ctx, logE, &domain.DownloadGitHubReleaseParam{ //nolint:wrapcheck
			RepoOwner: file.RepoOwner,
			RepoName:  file.RepoName,
			Version:   file.Version,
			Asset:     file.Asset,
			Private:   file.Private,
		})
	case config.PkgInfoTypeGitHubContent:
		file, err := downloader.ghContent.DownloadGitHubContentFile(ctx, logE, &domain.GitHubContentFileParam{
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
		return downloader.getReadCloserFromGitHubArchive(ctx, file)
	case config.PkgInfoTypeHTTP:
		rc, code, err := downloader.http.Download(ctx, file.URL)
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
