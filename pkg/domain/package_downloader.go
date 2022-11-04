package domain

import (
	"context"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type PackageDownloader interface {
	GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry, rt *runtime.Runtime) (io.ReadCloser, int64, error)
}

type MockPackageDownloader struct {
	Body string
	Code int64
	Err  error
}

func (dl *MockPackageDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry, rt *runtime.Runtime) (io.ReadCloser, int64, error) {
	if dl.Err == nil {
		return io.NopCloser(strings.NewReader(dl.Body)), dl.Code, nil
	}
	return nil, dl.Code, dl.Err
}
