package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v44/github"
)

var (
	errReleaseNotFound = errors.New("release isn't found")
	errAssetNotFound   = errors.New("asset isn't found")
	errContentNotFound = errors.New("content isn't found")
)

type mockRepositoryService struct {
	releases []*github.RepositoryRelease
	content  *github.RepositoryContent
	asset    string
}

func NewMock(releases []*github.RepositoryRelease, content *github.RepositoryContent, asset string) RepositoryService {
	return &mockRepositoryService{
		releases: releases,
		content:  content,
		asset:    asset,
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
