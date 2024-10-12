package generate

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, logE *logrus.Entry, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	ListReleases(ctx context.Context, logE *logrus.Entry, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	ListTags(ctx context.Context, logE *logrus.Entry, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
}

type Outputter interface {
	Output(param *output.Param) error
}
