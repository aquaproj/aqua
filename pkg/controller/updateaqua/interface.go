package updateaqua

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

type AquaInstaller interface {
	InstallAqua(ctx context.Context, logger *slog.Logger, version string) error
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}
