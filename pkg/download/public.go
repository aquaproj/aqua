package download

import (
	"context"
	"io"

	"github.com/clivm/clivm/pkg/config"
	"github.com/clivm/clivm/pkg/github"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type PackageDownloader interface {
	GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry) (io.ReadCloser, error)
}

func NewPackageDownloader(gh github.RepositoryService, rt *runtime.Runtime, httpDownloader HTTPDownloader) PackageDownloader {
	return &pkgDownloader{
		github:  gh,
		runtime: rt,
		http:    httpDownloader,
	}
}

type RegistryDownloader interface {
	GetGitHubContentFile(ctx context.Context, repoOwner, repoName, ref, path string, logE *logrus.Entry) ([]byte, error)
}

func NewRegistryDownloader(gh github.RepositoryService, httpDownloader HTTPDownloader) RegistryDownloader {
	return &registryDownloader{
		github: gh,
		http:   httpDownloader,
	}
}
