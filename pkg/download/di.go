package download

import (
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/log"
)

func NewPackageDownloader(gh github.RepositoryService, logger *log.Logger) PackageDownloader {
	return &PkgDownloader{
		GitHubRepositoryService: gh,
		logger:                  logger,
	}
}
