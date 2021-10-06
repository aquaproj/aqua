package controller

import (
	"fmt"
	"path/filepath"
)

type GitHubArchivePackageInfo struct {
	Name               string
	Link               string
	Description        string
	VersionConstraints *VersionConstraints
	VersionOverrides   []*GitHubArchiveVersionOverride
	Files              []*File `validate:"dive"`

	RepoOwner string
	RepoName  string
}

func (pkgInfo *GitHubArchivePackageInfo) SetVersion(v string) (PackageInfo, error) {
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
			return overrideGitHubArchivePackageInfo(pkgInfo, vo), nil
		}
	}
	return pkgInfo, nil
}

func overrideGitHubArchivePackageInfo(base *GitHubArchivePackageInfo, vo *GitHubArchiveVersionOverride) *GitHubArchivePackageInfo {
	p := &GitHubArchivePackageInfo{
		Name:        base.Name,
		Link:        base.Link,
		Description: base.Description,
		Files:       base.Files,
		RepoOwner:   base.RepoOwner,
		RepoName:    base.RepoName,
	}
	if vo.Files != nil {
		p.Files = vo.Files
	}
	return p
}

type GitHubArchiveVersionOverride struct {
	VersionConstraints *VersionConstraints
	Files              []*File `validate:"dive"`
}

func (pkgInfo *GitHubArchivePackageInfo) GetName() string {
	if pkgInfo.Name != "" {
		return pkgInfo.Name
	}
	return pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
}

func (pkgInfo *GitHubArchivePackageInfo) GetType() string {
	return pkgInfoTypeGitHubArchive
}

func (pkgInfo *GitHubArchivePackageInfo) GetLink() string {
	if pkgInfo.Link != "" {
		return pkgInfo.Link
	}
	return "https://github.com/" + pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
}

func (pkgInfo *GitHubArchivePackageInfo) GetDescription() string {
	return pkgInfo.Description
}

func (pkgInfo *GitHubArchivePackageInfo) GetFormat() string {
	return "tar.gz"
}

func (pkgInfo *GitHubArchivePackageInfo) GetReplacements() map[string]string {
	return nil
}

func (pkgInfo *GitHubArchivePackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
}

func (pkgInfo *GitHubArchivePackageInfo) GetFiles() []*File {
	if len(pkgInfo.Files) != 0 {
		return pkgInfo.Files
	}
	return []*File{
		{
			Name: pkgInfo.RepoName,
		},
	}
}

func (pkgInfo *GitHubArchivePackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
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

func (pkgInfo *GitHubArchivePackageInfo) RenderAsset(pkg *Package) (string, error) {
	return "", nil
}
