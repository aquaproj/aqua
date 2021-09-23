package controller

import (
	"fmt"
	"path/filepath"
	"runtime"
)

type GitHubReleasePackageInfo struct {
	Name               string
	Format             string
	Link               string
	Description        string
	Files              []*File `validate:"dive"`
	Replacements       map[string]string
	FormatOverrides    []*FormatOverride
	VersionConstraints *VersionConstraints
	VersionOverrides   []*GitHubReleaseVersionOverride

	RepoOwner string
	RepoName  string
	Asset     *Template `validate:"required"`
}

func (pkgInfo *GitHubReleasePackageInfo) SetVersion(v string) (PackageInfo, error) {
	if pkgInfo.VersionConstraints == nil {
		return pkgInfo, nil
	}
	a, err := pkgInfo.VersionConstraints.Check(v)
	if err != nil {
		return nil, err
	}
	if a {
		return pkgInfo, nil
	}
	for _, vo := range pkgInfo.VersionOverrides {
		a, err := vo.VersionConstraints.Check(v)
		if err != nil {
			return nil, err
		}
		if a {
			return overrideGitHubReleasePackageInfo(pkgInfo, vo), nil
		}
	}
	return pkgInfo, nil
}

func overrideGitHubReleasePackageInfo(base *GitHubReleasePackageInfo, vo *GitHubReleaseVersionOverride) *GitHubReleasePackageInfo {
	p := &GitHubReleasePackageInfo{
		Name:            base.Name,
		Format:          base.Format,
		Link:            base.Link,
		Description:     base.Description,
		Files:           base.Files,
		Replacements:    base.Replacements,
		FormatOverrides: base.FormatOverrides,
		RepoOwner:       base.RepoOwner,
		RepoName:        base.RepoName,
		Asset:           base.Asset,
	}
	if vo.Asset != nil {
		p.Asset = vo.Asset
	}
	if vo.Files != nil {
		p.Files = vo.Files
	}
	if vo.Format != "" {
		p.Format = vo.Format
	}
	if vo.FormatOverrides != nil {
		p.FormatOverrides = vo.FormatOverrides
	}
	if vo.Replacements != nil {
		p.Replacements = vo.Replacements
	}
	return p
}

type GitHubReleaseVersionOverride struct {
	VersionConstraints *VersionConstraints
	Asset              *Template `validate:"required"`
	Files              []*File   `validate:"dive"`
	Format             string
	FormatOverrides    []*FormatOverride
	Replacements       map[string]string
}

func (pkgInfo *GitHubReleasePackageInfo) GetName() string {
	if pkgInfo.Name != "" {
		return pkgInfo.Name
	}
	return pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
}

func (pkgInfo *GitHubReleasePackageInfo) GetType() string {
	return pkgInfoTypeGitHubRelease
}

func (pkgInfo *GitHubReleasePackageInfo) GetLink() string {
	if pkgInfo.Link != "" {
		return pkgInfo.Link
	}
	return "https://github.com/" + pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
}

func (pkgInfo *GitHubReleasePackageInfo) GetDescription() string {
	return pkgInfo.Description
}

func (pkgInfo *GitHubReleasePackageInfo) GetFormat() string {
	for _, arcTypeOverride := range pkgInfo.FormatOverrides {
		if arcTypeOverride.GOOS == runtime.GOOS {
			return arcTypeOverride.Format
		}
	}
	return pkgInfo.Format
}

func (pkgInfo *GitHubReleasePackageInfo) GetReplacements() map[string]string {
	return pkgInfo.Replacements
}

func (pkgInfo *GitHubReleasePackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
}

func (pkgInfo *GitHubReleasePackageInfo) GetFiles() []*File {
	if len(pkgInfo.Files) != 0 {
		return pkgInfo.Files
	}
	return []*File{
		{
			Name: pkgInfo.RepoName,
		},
	}
}

func (pkgInfo *GitHubReleasePackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if isUnarchived(pkgInfo.GetFormat(), assetName) {
		return assetName, nil
	}
	if file.Src == nil {
		return file.Name, nil
	}
	src, err := file.RenderSrc(pkg, pkgInfo)
	if err != nil {
		return "", fmt.Errorf("render the template file.src: %w", err)
	}
	return src, nil
}

func (pkgInfo *GitHubReleasePackageInfo) RenderAsset(pkg *Package) (string, error) {
	return pkgInfo.Asset.Execute(map[string]interface{}{
		"Version": pkg.Version,
		"GOOS":    runtime.GOOS,
		"GOARCH":  runtime.GOARCH,
		"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":    replace(runtime.GOARCH, pkgInfo.GetReplacements()),
		"Format":  pkgInfo.GetFormat(),
	})
}
