package github

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/google/go-github/v66/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type (
	ReleaseAsset                = github.ReleaseAsset
	ListOptions                 = github.ListOptions
	RepositoryRelease           = github.RepositoryRelease
	RepositoriesServiceImpl     = github.RepositoriesService
	Repository                  = github.Repository
	RepositoryContentGetOptions = github.RepositoryContentGetOptions
	RepositoryContent           = github.RepositoryContent
	Response                    = github.Response
	RepositoryTag               = github.RepositoryTag
	ArchiveFormat               = github.ArchiveFormat
)

const Tarball = github.Tarball

type Option struct {
	Keyring bool
}

func New(ctx context.Context, opt *Option) *GitHub {
	token := getGitHubToken()
	if opt == nil || !opt.Keyring || token != "" {
		return &GitHub{
			repo:  github.NewClient(getHTTPClientForGitHub(ctx, token)).Repositories,
			mutex: &sync.RWMutex{},
		}
	}
	return &GitHub{
		mutex: &sync.RWMutex{},
	}
}

type RepositoriesService interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, maxRedirects int) (*url.URL, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
	DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	ListReleaseAssets(ctx context.Context, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error)
}

func getGitHubToken() string {
	if token := os.Getenv("AQUA_GITHUB_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

func getHTTPClientForGitHub(ctx context.Context, token string) *http.Client {
	if token == "" {
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
}

type GitHub struct {
	repo  RepositoriesService
	mutex *sync.RWMutex
}

func (g *GitHub) init(ctx context.Context, logE *logrus.Entry) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.repo != nil {
		return
	}
	token, err := getTokenFromKeyring()
	if err != nil {
		logE.WithError(err).Warn("get a GitHub Access token from keyring")
		g.repo = github.NewClient(http.DefaultClient).Repositories
	}
	logE.Debug("got a GitHub Access token from keyring")
	g.repo = github.NewClient(getHTTPClientForGitHub(ctx, token)).Repositories
}

func (g *GitHub) GetArchiveLink(ctx context.Context, logE *logrus.Entry, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, maxRedirects int) (*url.URL, *github.Response, error) {
	g.init(ctx, logE)
	return g.repo.GetArchiveLink(ctx, owner, repo, archiveformat, opts, maxRedirects)
}

func (g *GitHub) GetReleaseByTag(ctx context.Context, logE *logrus.Entry, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error) {
	g.init(ctx, logE)
	return g.repo.GetReleaseByTag(ctx, owner, repoName, version)
}

func (g *GitHub) DownloadReleaseAsset(ctx context.Context, logE *logrus.Entry, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error) {
	g.init(ctx, logE)
	return g.repo.DownloadReleaseAsset(ctx, owner, repoName, assetID, httpClient)
}

func (g *GitHub) DownloadContents(ctx context.Context, logE *logrus.Entry, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error) {
	g.init(ctx, logE)
	return g.repo.DownloadContents(ctx, owner, repo, filepath, opts)
}

func (g *GitHub) GetLatestRelease(ctx context.Context, logE *logrus.Entry, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error) {
	g.init(ctx, logE)
	return g.repo.GetLatestRelease(ctx, repoOwner, repoName)
}

func (g *GitHub) ListReleases(ctx context.Context, logE *logrus.Entry, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error) {
	g.init(ctx, logE)
	return g.repo.ListReleases(ctx, owner, repo, opts)
}

func (g *GitHub) ListTags(ctx context.Context, logE *logrus.Entry, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
	g.init(ctx, logE)
	return g.repo.ListTags(ctx, owner, repo, opts)
}

func (g *GitHub) Get(ctx context.Context, logE *logrus.Entry, owner, repo string) (*github.Repository, *github.Response, error) {
	g.init(ctx, logE)
	return g.repo.Get(ctx, owner, repo)
}

func (g *GitHub) ListReleaseAssets(ctx context.Context, logE *logrus.Entry, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error) {
	g.init(ctx, logE)
	return g.repo.ListReleaseAssets(ctx, owner, repo, id, opts)
}
