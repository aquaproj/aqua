package config

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/runtime"
)

func (cpkg *Package) GetChecksumID(rt *runtime.Runtime) (string, error) {
	assetName, err := cpkg.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGo:
		return filepath.Join(pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return filepath.Join(pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
		uS, err := cpkg.RenderURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return filepath.Join(pkgInfo.GetType(), u.Host, u.Path), nil
	}
	return "", nil
}
