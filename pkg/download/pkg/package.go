package pkg

// import (
// 	"context"
// 	"io"
//
// 	"github.com/aquaproj/aqua/pkg/config"
// 	"github.com/aquaproj/aqua/pkg/download"
// 	"github.com/aquaproj/aqua/pkg/runtime"
// 	"github.com/sirupsen/logrus"
// )
//
// type PackageDownloader struct {
// 	runtime *runtime.Runtime
// 	dl      *download.Downloader
// }
//
// func NewPackageDownloader(rt *runtime.Runtime, dl *download.Downloader) *PackageDownloader {
// 	return &PackageDownloader{
// 		runtime: rt,
// 		dl:      dl,
// 	}
// }
//
// func (downloader *PackageDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry, rt *runtime.Runtime) (io.ReadCloser, int64, error) {
// 	if rt == nil {
// 		rt = downloader.runtime
// 	}
// 	file, err := download.ConvertPackageToFile(pkg, assetName, rt)
// 	if err != nil {
// 		return nil, 0, err //nolint:wrapcheck
// 	}
// 	return downloader.dl.GetReadCloser(ctx, file, logE) //nolint:wrapcheck
// }
