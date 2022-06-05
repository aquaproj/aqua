package download

import (
	"context"

	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
)

type RegistryDownloader interface {
	GetGitHubContentFile(ctx context.Context, repoOwner, repoName, ref, path string, logE *logrus.Entry) ([]byte, error)
}

func NewRegistryDownloader(gh github.RepositoryService, httpDownloader HTTPDownloader) RegistryDownloader {
	return &registryDownloader{
		github: gh,
		http:   httpDownloader,
	}
}
