package download

import (
	"context"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type PackageDownloader struct {
	github  domain.RepositoriesService
	runtime *runtime.Runtime
	dl      *Downloader
}

func NewPackageDownloader(gh domain.RepositoriesService, rt *runtime.Runtime, dl *Downloader) *PackageDownloader {
	return &PackageDownloader{
		github:  gh,
		runtime: rt,
		dl:      dl,
	}
}

func (downloader *PackageDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry, rt *runtime.Runtime) (io.ReadCloser, int64, error) {
	if rt == nil {
		rt = downloader.runtime
	}
	file, err := ConvertPackageToFile(pkg, assetName, rt)
	if err != nil {
		return nil, 0, err
	}
	return downloader.dl.GetReadCloser(ctx, file, logE)
}
