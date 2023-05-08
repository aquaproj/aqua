package initcmd

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
}
