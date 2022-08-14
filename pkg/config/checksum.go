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
		uS, err := cpkg.RenderURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return path.Join(pkgInfo.GetType(), u.Host, u.Path), nil
	}
	return "", nil
}

func (cpkg *Package) getRuntimeFromAsset(asset string) (*runtime.Runtime, error) {
	rts, err := runtime.GetRuntimesFromEnvs(cpkg.PackageInfo.SupportedEnvs)
	if err != nil {
		return nil, fmt.Errorf("get supported runtimes from supported_envs: %w", err)
	}
	for _, rt := range rts {
		a, err := cpkg.RenderAsset(rt)
		if err != nil {
			return nil, err
		}
		if a == asset {
			return rt, nil
		}
	}
	return nil, nil //nolint:nilnil
}

func (cpkg *Package) GetChecksumIDFromAsset(asset string) (string, error) {
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGo:
		return path.Join(pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return path.Join(pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, asset), nil
	case PkgInfoTypeHTTP:
		rt, err := cpkg.getRuntimeFromAsset(asset)
		if err != nil {
			return "", fmt.Errorf("get a runtime from an asset: %w", err)
		}
		if rt == nil {
			return "", nil
		}
		return cpkg.GetChecksumID(rt)
	}
	return "", nil
}

func (cpkg *Package) RenderChecksumFileName(rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	switch pkgInfo.Checksum.Type { //nolint:gocritic
	case PkgInfoTypeGitHubRelease:
		asset, err := cpkg.RenderAsset(rt)
		if err != nil {
			return "", err
		}
		return cpkg.renderChecksumFile(asset, rt)
	}
	return "", errUnknownChecksumFileType
}

func (cpkg *Package) RenderChecksumURL(rt *runtime.Runtime) (string, error) {
	return cpkg.renderTemplateString(cpkg.PackageInfo.Checksum.URL, rt)
}

func (cpkg *Package) RenderChecksumFileID(rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	switch pkgInfo.Checksum.Type {
	case PkgInfoTypeGitHubRelease:
		return cpkg.RenderChecksumFileName(rt)
	case PkgInfoTypeHTTP:
		return cpkg.RenderChecksumURL(rt)
	}
	return "", errUnknownChecksumFileType
}
