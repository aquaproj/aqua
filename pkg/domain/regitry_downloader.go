package domain

import (
	"context"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
)

type PackageDownloader interface {
	GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry) (io.ReadCloser, int64, error)
}

type RegistryDownloader interface {
	GetGitHubContentFile(ctx context.Context, repoOwner, repoName, ref, path string, logE *logrus.Entry) ([]byte, error)
}
