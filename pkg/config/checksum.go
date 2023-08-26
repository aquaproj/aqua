package config

import (
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

var errUnknownChecksumFileType = errors.New("unknown checksum type")

func (p *Package) ChecksumID(rt *runtime.Runtime) (string, error) {
	assetName, err := p.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	pkgInfo := p.PackageInfo
	pkg := p.Package
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return path.Join(pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return path.Join(pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
		uS, err := p.RenderURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return path.Join(pkgInfo.Type, u.Host, u.Path), nil
	}
	return "", nil
}

func (p *Package) getRuntimeFromAsset(asset string) (*runtime.Runtime, error) {
	rts, err := runtime.GetRuntimesFromEnvs(p.PackageInfo.SupportedEnvs)
	if err != nil {
		return nil, fmt.Errorf("get supported runtimes from supported_envs: %w", err)
	}
	for _, rt := range rts {
		a, err := p.RenderAsset(rt)
		if err != nil {
			return nil, err
		}
		if a == asset {
			return rt, nil
		}
	}
	return nil, nil //nolint:nilnil
}

func (p *Package) ChecksumIDFromAsset(asset string) (string, error) {
	pkgInfo := p.PackageInfo
	pkg := p.Package
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return path.Join(pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		return path.Join(pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, asset), nil
	case PkgInfoTypeHTTP:
		rt, err := p.getRuntimeFromAsset(asset)
		if err != nil {
			return "", fmt.Errorf("get a runtime from an asset: %w", err)
		}
		if rt == nil {
			return "", nil
		}
		return p.ChecksumID(rt)
	}
	return "", nil
}

func (p *Package) RenderChecksumFileName(rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	switch pkgInfo.Checksum.Type { //nolint:gocritic
	case PkgInfoTypeGitHubRelease:
		asset, err := p.RenderAsset(rt)
		if err != nil {
			return "", err
		}
		return p.renderChecksumFile(asset, rt)
	}
	return "", errUnknownChecksumFileType
}

func (p *Package) RenderChecksumURL(rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	pkg := p.Package
	replacements := pkgInfo.GetChecksumReplacements()
	m := map[string]interface{}{
		"Version": pkg.Version,
		"SemVer":  p.semVer(),
		"GOOS":    rt.GOOS,
		"GOARCH":  rt.GOARCH,
		"OS":      replace(rt.GOOS, replacements),
		"Arch":    getArch(pkgInfo.Rosetta2, replacements, rt),
		"Format":  pkgInfo.GetFormat(),
	}
	if pkgInfo.Type == "http" {
		u, err := p.RenderURL(rt)
		if err != nil {
			return "", err
		}
		m["AssetURL"] = u
	}

	return template.Execute(p.PackageInfo.Checksum.URL, m) //nolint:wrapcheck
}

func (p *Package) RenderChecksumFileID(rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	switch pkgInfo.Checksum.Type {
	case PkgInfoTypeGitHubRelease:
		return p.RenderChecksumFileName(rt)
	case PkgInfoTypeHTTP:
		return p.RenderChecksumURL(rt)
	}
	return "", errUnknownChecksumFileType
}
