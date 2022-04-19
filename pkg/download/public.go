package download

import (
	"context"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type PackageDownloader interface {
	GetReadCloser(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, assetName string, logE *logrus.Entry) (io.ReadCloser, error)
}

func NewPackageDownloader(gh github.RepositoryService, rt *runtime.Runtime) PackageDownloader {
	return &pkgDownloader{
		github:  gh,
		runtime: rt,
	}
}

type RegistryDownloader interface {
	GetGitHubContentFile(ctx context.Context, repoOwner, repoName, ref, path string, logE *logrus.Entry) ([]byte, error)
}

func NewRegistryDownloader(gh github.RepositoryService) RegistryDownloader {
	return &registryDownloader{
		github: gh,
	}
}
