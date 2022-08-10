package config

import (
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/aquaproj/aqua/pkg/runtime"
)

var errUnknownChecksumFileType = errors.New("unknown checksum type")

func (cpkg *Package) GetChecksumID(rt *runtime.Runtime) (string, error) {
	assetName, err := cpkg.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGo:
		return path.Join(pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return path.Join(pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
		uS, err := cpkg.RenderChecksumURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		asset, err := cpkg.renderAsset(rt)
		if err != nil {
			return "", fmt.Errorf("get the asset name: %w", err)
		}
		return path.Join("http", u.Host, u.Path, asset), nil
	}
	return "", nil
}

func (cpkg *Package) GetChecksumIDFromAsset(rt *runtime.Runtime, asset string) (string, error) {
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGo:
		return path.Join(pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return path.Join(pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, asset), nil
	case PkgInfoTypeHTTP:
		uS, err := cpkg.RenderChecksumURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return path.Join("http", u.Host, u.Path, asset), nil
	}
	return "", nil
}

func (cpkg *Package) RenderChecksumFileName(rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	switch pkgInfo.Checksum.Type {
	case PkgInfoTypeGitHubRelease, "github_release_multifile":
		return cpkg.renderTemplateString(pkgInfo.Checksum.Asset, rt)
	}
	return "", errUnknownChecksumFileType
}

func (cpkg *Package) RenderChecksumURL(rt *runtime.Runtime) (string, error) {
	return cpkg.renderTemplateString(cpkg.PackageInfo.Checksum.URL, rt)
}
