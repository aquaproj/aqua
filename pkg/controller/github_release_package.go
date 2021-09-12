package controller

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
)

type GitHubReleasePackageInfo struct {
	Name                 string `validate:"required"`
	ArchiveType          string
	Link                 string
	Description          string
	Files                []*File `validate:"required,dive"`
	Replacements         map[string]string
	ArchiveTypeOverrides []*ArchiveTypeOverride

	RepoOwner string
	RepoName  string
	Asset     *text.Template `validate:"required"`
}

func (pkgInfo *GitHubReleasePackageInfo) GetName() string {
	return pkgInfo.Name
}

func (pkgInfo *GitHubReleasePackageInfo) GetType() string {
	return pkgInfoTypeGitHubRelease
}

func (pkgInfo *GitHubReleasePackageInfo) GetLink() string {
	return pkgInfo.Link
}

func (pkgInfo *GitHubReleasePackageInfo) GetDescription() string {
	return pkgInfo.Description
}

func (pkgInfo *GitHubReleasePackageInfo) GetArchiveType() string {
	for _, arcTypeOverride := range pkgInfo.ArchiveTypeOverrides {
		if arcTypeOverride.GOOS == runtime.GOOS {
			return arcTypeOverride.ArchiveType
		}
	}
	return pkgInfo.ArchiveType
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
		"Version":     pkg.Version,
		"GOOS":        runtime.GOOS,
		"GOARCH":      runtime.GOARCH,
		"OS":          replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":        replace(runtime.GOARCH, pkgInfo.GetReplacements()),
		"ArchiveType": pkgInfo.GetArchiveType(),

		// DEPRECATED: don't use these variables
		"Package":     pkg,
		"PackageInfo": pkgInfo,
	})
}
