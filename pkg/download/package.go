package download

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type PackageDownloader struct {
	github    domain.RepositoriesService
	runtime   *runtime.Runtime
	http      HTTPDownloader
	ghContent domain.GitHubContentFileDownloader
}

func NewPackageDownloader(gh domain.RepositoriesService, rt *runtime.Runtime, httpDownloader HTTPDownloader) *PackageDownloader {
	return &PackageDownloader{
		github:    gh,
		runtime:   rt,
		http:      httpDownloader,
		ghContent: NewGitHubContentFileDownloader(gh, httpDownloader),
	}
}

func (downloader *PackageDownloader) GetReadCloser(ctx context.Context, pkg *config.Package, assetName string, logE *logrus.Entry) (io.ReadCloser, int64, error) {
	pkgInfo := pkg.PackageInfo
	switch pkgInfo.GetType() {
	case config.PkgInfoTypeGitHubRelease:
		pkgInfo := pkg.PackageInfo
		return downloader.downloadFromGitHubRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, assetName, logE)
	case config.PkgInfoTypeGitHubContent:
		pkgInfo := pkg.PackageInfo
		file, err := downloader.ghContent.DownloadGitHubContentFile(ctx, logE, &domain.GitHubContentFileParam{
			RepoOwner: pkgInfo.RepoOwner,
			RepoName:  pkgInfo.RepoName,
			Ref:       pkg.Package.Version,
			Path:      assetName,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("download a package from GitHub Content: %w", err)
		}
		if file.ReadCloser != nil {
			return file.ReadCloser, 0, nil
		}
		return io.NopCloser(strings.NewReader(file.String)), 0, nil
	case config.PkgInfoTypeGitHubArchive, config.PkgInfoTypeGo:
		return downloader.getReadCloserFromGitHubArchive(ctx, pkg)
	case config.PkgInfoTypeHTTP:
		uS, err := pkg.RenderURL(downloader.runtime)
		if err != nil {
			return nil, 0, err //nolint:wrapcheck
		}
		return downloader.http.Download(ctx, uS) //nolint:wrapcheck
	default:
		return nil, 0, logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}
}
