package updateaqua

import (
	"context"

	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
)

type AquaInstaller interface {
	InstallAqua(ctx context.Context, logE *logrus.Entry, version string) error
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}
