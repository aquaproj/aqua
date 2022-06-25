package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v44/github"
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

type mockRepositoryService struct {
	releases []*github.RepositoryRelease
	content  *github.RepositoryContent
	repo     *github.Repository
	tags     []*github.RepositoryTag
	asset    string
	assets   []*github.ReleaseAsset
	url      *url.URL
}

func NewMock(releases []*github.RepositoryRelease, content *github.RepositoryContent, asset string, assets []*github.ReleaseAsset, url *url.URL) RepositoryService {
	return &mockRepositoryService{
		releases: releases,
		content:  content,
		asset:    asset,
		assets:   assets,
		url:      url,
	}
}

func (svc *mockRepositoryService) GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error) {
	if len(svc.releases) == 0 {
		return nil, nil, errReleaseNotFound
	}
	return svc.releases[0], nil, nil
}

func (svc *mockRepositoryService) GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	if svc.content == nil {
		return svc.content, nil, nil, errContentNotFound
	}
	return svc.content, nil, nil, nil
}

func (svc *mockRepositoryService) GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error) {
	if len(svc.releases) == 0 {
		return nil, nil, errReleaseNotFound
	}
	return svc.releases[0], nil, nil
}

func (svc *mockRepositoryService) DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error) {
	if svc.asset == "" {
		return nil, "", errAssetNotFound
	}
	return io.NopCloser(strings.NewReader(svc.asset)), "", nil
}

func (svc *mockRepositoryService) ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error) {
	if svc.releases == nil {
		return nil, nil, errReleaseNotFound
	}
	return svc.releases, nil, nil
}

func (svc *mockRepositoryService) ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
	if svc.tags == nil {
		return nil, nil, errTagNotFound
	}
	return svc.tags, nil, nil
}

func (svc *mockRepositoryService) GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error) {
	if svc.url == nil {
		return nil, nil, errGetTar
	}
	return svc.url, nil, nil
}

func (svc *mockRepositoryService) Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	if svc.repo == nil {
		return nil, nil, errGetRepo
	}
	return svc.repo, nil, nil
}

func (svc *mockRepositoryService) ListReleaseAssets(ctx context.Context, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error) {
	if svc.assets == nil {
		return nil, nil, errListAssets
	}
	return svc.assets, nil, nil
}
