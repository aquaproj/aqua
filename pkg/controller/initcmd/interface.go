package initcmd

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, logE *logrus.Entry, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
}
