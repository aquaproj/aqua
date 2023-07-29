package download

import (
	"errors"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func ConvertDownloadedFileToFile(file *registry.DownloadedFile, art *File, rt *runtime.Runtime, tplParam *template.Artifact) (*File, error) {
	f := &File{
		Type:      file.Type,
		RepoOwner: file.RepoOwner,
		RepoName:  file.RepoName,
		Version:   art.Version,
	}
	switch file.Type {
	case "github_release":
		if f.RepoOwner == "" {
			f.RepoOwner = art.RepoOwner
		}
		if f.RepoName == "" {
			f.RepoName = art.RepoName
		}
		if file.Asset == nil {
			return nil, errors.New("asset is required")
		}
		asset, err := template.Render(*file.Asset, tplParam, rt)
		if err != nil {
			return nil, fmt.Errorf("render an asset template: %w", err)
		}
		f.Asset = asset
		return f, nil
	case "http":
		if file.URL == nil {
			return nil, errors.New("url is required")
		}
		u, err := template.Render(*file.URL, tplParam, rt)
		if err != nil {
			return nil, fmt.Errorf("render a url template: %w", err)
		}
		f.URL = u
		return f, nil
	}
	return nil, logerr.WithFields(errors.New("invalid file type"), logrus.Fields{ //nolint:wrapcheck
		"file_type": file.Type,
	})
}

func ConvertPackageToFile(pkg *config.Package, assetName string, rt *runtime.Runtime) (*File, error) {
	pkgInfo := pkg.PackageInfo
	file := &File{
		Type:      pkgInfo.GetType(),
		RepoOwner: pkgInfo.RepoOwner,
		RepoName:  pkgInfo.RepoName,
		Version:   pkg.Package.Version,
		Private:   pkgInfo.Private,
	}
	switch pkgInfo.GetType() {
	case config.PkgInfoTypeGitHubRelease:
		file.Asset = assetName
		return file, nil
	case config.PkgInfoTypeGitHubContent:
		file.Path = assetName
		return file, nil
	case config.PkgInfoTypeGitHubArchive, config.PkgInfoTypeGoBuild:
		file.Type = config.PkgInfoTypeGitHubArchive
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
		Private:   rgst.Private,
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
