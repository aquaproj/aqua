package genrgst

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

type RepositoriesService interface {
	Get(ctx context.Context, logE *logrus.Entry, owner, repo string) (*github.Repository, *github.Response, error)
	GetLatestRelease(ctx context.Context, logE *logrus.Entry, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	GetReleaseByTag(ctx context.Context, logE *logrus.Entry, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
	ListReleaseAssets(ctx context.Context, logE *logrus.Entry, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error)
	ListReleases(ctx context.Context, logE *logrus.Entry, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
}
