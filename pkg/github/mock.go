package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v56/github"
)

var (
	errReleaseNotFound = errors.New("release isn't found")
	errTagNotFound     = errors.New("tag isn't found")
	errAssetNotFound   = errors.New("asset isn't found")
	errContentNotFound = errors.New("content isn't found")
	errGetTar          = errors.New("failed to get tar")
	errGetRepo         = errors.New("failed to get repo")
	errListAssets      = errors.New("failed to list assets")
)

type MockRepositoriesService struct {
	Releases []*github.RepositoryRelease
	Content  *github.RepositoryContent
	Repo     *github.Repository
	Tags     []*github.RepositoryTag
	Asset    string
	Assets   []*github.ReleaseAsset
	URL      *url.URL
}

func (m *MockRepositoriesService) GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error) {
	if len(m.Releases) == 0 {
		return nil, nil, errReleaseNotFound
	}
	return m.Releases[0], nil, nil
}

func (m *MockRepositoriesService) GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	if m.Content == nil {
		return m.Content, nil, nil, errContentNotFound
	}
	return m.Content, nil, nil, nil
}

func (m *MockRepositoriesService) GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error) {
	if len(m.Releases) == 0 {
		return nil, nil, errReleaseNotFound
	}
	return m.Releases[0], nil, nil
}

func (m *MockRepositoriesService) DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error) {
	if m.Asset == "" {
		return nil, "", errAssetNotFound
	}
	return io.NopCloser(strings.NewReader(m.Asset)), "", nil
}

func (m *MockRepositoriesService) ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error) {
	if m.Releases == nil {
		return nil, nil, errReleaseNotFound
	}
	return m.Releases, nil, nil
}

func (m *MockRepositoriesService) ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
	if m.Tags == nil {
		return nil, nil, errTagNotFound
	}
	return m.Tags, nil, nil
}

func (m *MockRepositoriesService) GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error) {
	if m.URL == nil {
		return nil, nil, errGetTar
	}
	return m.URL, nil, nil
}

func (m *MockRepositoriesService) Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	if m.Repo == nil {
		return nil, nil, errGetRepo
	}
	return m.Repo, nil, nil
}

func (m *MockRepositoriesService) ListReleaseAssets(ctx context.Context, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error) {
	if m.Assets == nil {
		return nil, nil, errListAssets
	}
	return m.Assets, nil, nil
}
