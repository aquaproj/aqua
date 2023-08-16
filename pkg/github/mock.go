package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v54/github"
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

func (svc *MockRepositoriesService) GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error) {
	if len(svc.Releases) == 0 {
		return nil, nil, errReleaseNotFound
	}
	return svc.Releases[0], nil, nil
}

func (svc *MockRepositoriesService) GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	if svc.Content == nil {
		return svc.Content, nil, nil, errContentNotFound
	}
	return svc.Content, nil, nil, nil
}

func (svc *MockRepositoriesService) GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error) {
	if len(svc.Releases) == 0 {
		return nil, nil, errReleaseNotFound
	}
	return svc.Releases[0], nil, nil
}

func (svc *MockRepositoriesService) DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error) {
	if svc.Asset == "" {
		return nil, "", errAssetNotFound
	}
	return io.NopCloser(strings.NewReader(svc.Asset)), "", nil
}

func (svc *MockRepositoriesService) ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error) {
	if svc.Releases == nil {
		return nil, nil, errReleaseNotFound
	}
	return svc.Releases, nil, nil
}

func (svc *MockRepositoriesService) ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
	if svc.Tags == nil {
		return nil, nil, errTagNotFound
	}
	return svc.Tags, nil, nil
}

func (svc *MockRepositoriesService) GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error) {
	if svc.URL == nil {
		return nil, nil, errGetTar
	}
	return svc.URL, nil, nil
}

func (svc *MockRepositoriesService) Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	if svc.Repo == nil {
		return nil, nil, errGetRepo
	}
	return svc.Repo, nil, nil
}

func (svc *MockRepositoriesService) ListReleaseAssets(ctx context.Context, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error) {
	if svc.Assets == nil {
		return nil, nil, errListAssets
	}
	return svc.Assets, nil, nil
}
