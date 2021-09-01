package controller

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
)

type GitHubReleasePackageInfo struct {
	Name        string  `validate:"required"`
	ArchiveType string  `yaml:"archive_type"`
	Files       []*File `validate:"required,dive"`

	RepoOwner string         `yaml:"repo_owner" validate:"required"`
	RepoName  string         `yaml:"repo_name" validate:"required"`
	Asset     *text.Template `validate:"required"`
}

func (pkgInfo *GitHubReleasePackageInfo) GetName() string {
	return pkgInfo.Name
}

func (pkgInfo *GitHubReleasePackageInfo) GetType() string {
	return pkgInfoTypeGitHubRelease
}

func (pkgInfo *GitHubReleasePackageInfo) GetArchiveType() string {
	return pkgInfo.ArchiveType
}

func (pkgInfo *GitHubReleasePackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
}

func (pkgInfo *GitHubReleasePackageInfo) GetFiles() []*File {
	return pkgInfo.Files
}

func (pkgInfo *GitHubReleasePackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if isUnarchived(pkgInfo.GetArchiveType(), assetName) {
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
	return pkgInfo.Asset.Execute(map[string]interface{}{ //nolint:wrapcheck
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
	})
}
