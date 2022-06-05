package github

import (
	"context"
	"net/url"

	"github.com/google/go-github/v44/github"
)

const Tarball = github.Tarball

type (
	RepositoryContentGetOptions = github.RepositoryContentGetOptions
	ReleaseAsset                = github.ReleaseAsset
)

type ArchiveClient interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error)
}

func NewArchiveClient(repo RepositoryService) ArchiveClient {
	return repo
}
