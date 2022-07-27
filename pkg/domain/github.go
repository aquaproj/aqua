package domain

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
)

type RepositoriesService interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
	GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
}

type GitHubContentFileParam struct {
	RepoOwner string
	RepoName  string
	Ref       string
	Path      string
}

type GitHubContentFile struct {
	ReadCloser io.ReadCloser
	String     string
}

type GitHubContentFileDownloader interface {
	DownloadGitHubContentFile(ctx context.Context, logE *logrus.Entry, param *GitHubContentFileParam) (*GitHubContentFile, error)
}
