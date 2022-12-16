package download

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type File struct {
	Type      string `validate:"required" json:"type" jsonschema:"enum=github_release,enum=http"`
	RepoOwner string `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	RepoName  string `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	Version   string `yaml:",omitempty" json:"version,omitempty"`
	Asset     string `json:"asset,omitempty" yaml:",omitempty"`
	URL       string `json:"url,omitempty" yaml:",omitempty"`
	Path      string `json:"path,omitempty" yaml:",omitempty"`
}

type Downloader struct {
	github    domain.RepositoriesService
	http      HTTPDownloader
	ghContent domain.GitHubContentFileDownloader
	ghRelease domain.GitHubReleaseDownloader
}

func NewDownloader(gh domain.RepositoriesService, httpDownloader HTTPDownloader) *Downloader {
	return &Downloader{
		github:    gh,
		http:      httpDownloader,
		ghContent: NewGitHubContentFileDownloader(gh, httpDownloader),
		ghRelease: NewGitHubReleaseDownloader(gh, httpDownloader),
	}
}

func ConvertPackageToFile(pkg *config.Package, assetName string, rt *runtime.Runtime) (*File, error) {
	pkgInfo := pkg.PackageInfo
	file := &File{
		Type:      pkgInfo.GetType(),
		RepoOwner: pkgInfo.RepoOwner,
		RepoName:  pkgInfo.RepoName,
		Version:   pkg.Package.Version,
	}
	switch pkgInfo.GetType() {
	case config.PkgInfoTypeGitHubRelease:
		file.Asset = assetName
		return file, nil
	case config.PkgInfoTypeGitHubContent:
		file.Path = assetName
		return file, nil
	case config.PkgInfoTypeGitHubArchive, config.PkgInfoTypeGo:
		return file, nil
	case config.PkgInfoTypeHTTP:
		uS, err := pkg.RenderURL(rt)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		file.URL = uS
		return file, nil
	default:
		return nil, logerr.WithFields(domain.ErrInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"package_type": pkgInfo.GetType(),
		})
	}
}

func ConvertRegistryToFile(rgst *aqua.Registry) (*File, error) {
	file := &File{
		Type:      rgst.Type,
		RepoOwner: rgst.RepoOwner,
		RepoName:  rgst.RepoName,
		Version:   rgst.Ref,
	}
	switch rgst.Type {
	case config.PkgInfoTypeGitHubContent:
		file.Path = rgst.Path
		return file, nil
	default:
		return nil, logerr.WithFields(domain.ErrInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"registry_type": rgst.Type,
		})
	}
}

func (downloader *Downloader) GetReadCloser(ctx context.Context, file *File, logE *logrus.Entry) (io.ReadCloser, int64, error) {
	switch file.Type {
	case config.PkgInfoTypeGitHubRelease:
		return downloader.ghRelease.DownloadGitHubRelease(ctx, logE, &domain.DownloadGitHubReleaseParam{ //nolint:wrapcheck
			RepoOwner: file.RepoOwner,
			RepoName:  file.RepoName,
			Version:   file.Version,
			Asset:     file.Asset,
		})
	case config.PkgInfoTypeGitHubContent:
		file, err := downloader.ghContent.DownloadGitHubContentFile(ctx, logE, &domain.GitHubContentFileParam{
			RepoOwner: file.RepoOwner,
			RepoName:  file.RepoName,
			Ref:       file.Version,
			Path:      file.Path,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("download a package from GitHub Content: %w", err)
		}
		if file.ReadCloser != nil {
			return file.ReadCloser, 0, nil
		}
		return io.NopCloser(strings.NewReader(file.String)), 0, nil
	case config.PkgInfoTypeGitHubArchive:
		return downloader.getReadCloserFromGitHubArchive(ctx, file)
	case config.PkgInfoTypeHTTP:
		rc, code, err := downloader.http.Download(ctx, file.URL)
		if err != nil {
			return rc, code, fmt.Errorf("download a package: %w", logerr.WithFields(err, logrus.Fields{
				"download_url": file.URL,
			}))
		}
		return rc, code, nil
	default:
		return nil, 0, logerr.WithFields(errInvalidPackageType, logrus.Fields{ //nolint:wrapcheck
			"file_type": file.Type,
		})
	}
}
