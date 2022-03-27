package download

import (
	"context"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/log"
)

type PackageDownloader interface {
	GetReadCloser(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string) (io.ReadCloser, error)
}

func NewPackageDownloader(gh github.RepositoryService, logger *log.Logger) PackageDownloader {
	return &pkgDownloader{
		github: gh,
		logger: logger,
	}
}

type RegistryDownloader interface {
	GetGitHubContentFile(ctx context.Context, repoOwner, repoName, ref, path string) ([]byte, error)
}

func NewRegistryDownloader(gh github.RepositoryService, logger *log.Logger) RegistryDownloader {
	return &registryDownloader{
		github: gh,
		logger: logger,
	}
}
