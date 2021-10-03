package controller

import (
	"fmt"
	"path/filepath"
	"runtime"
)

type GitHubContentPackageInfo struct {
	Name               string
	Format             string
	Link               string
	Description        string
	Replacements       map[string]string
	FormatOverrides    []*FormatOverride
	VersionConstraints *VersionConstraints
	VersionOverrides   []*GitHubContentVersionOverride
	Files              []*File `validate:"dive"`

	RepoOwner string
	RepoName  string
	Path      *Template `validate:"required"`
}

func (pkgInfo *GitHubContentPackageInfo) SetVersion(v string) (PackageInfo, error) {
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
			return overrideGitHubContentPackageInfo(pkgInfo, vo), nil
		}
	}
	return pkgInfo, nil
}

func overrideGitHubContentPackageInfo(base *GitHubContentPackageInfo, vo *GitHubContentVersionOverride) *GitHubContentPackageInfo {
	p := &GitHubContentPackageInfo{
		Name:            base.Name,
		Format:          base.Format,
		Link:            base.Link,
		Description:     base.Description,
		Files:           base.Files,
		Replacements:    base.Replacements,
		FormatOverrides: base.FormatOverrides,
		RepoOwner:       base.RepoOwner,
		RepoName:        base.RepoName,
		Path:            base.Path,
	}
	if vo.Path != nil {
		p.Path = vo.Path
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

type GitHubContentVersionOverride struct {
	VersionConstraints *VersionConstraints
	Path               *Template `validate:"required"`
	Files              []*File   `validate:"dive"`
	Format             string
	FormatOverrides    []*FormatOverride
	Replacements       map[string]string
}

func (pkgInfo *GitHubContentPackageInfo) GetName() string {
	if pkgInfo.Name != "" {
		return pkgInfo.Name
	}
	return pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
}

func (pkgInfo *GitHubContentPackageInfo) GetType() string {
	return pkgInfoTypeGitHubContent
}

func (pkgInfo *GitHubContentPackageInfo) GetLink() string {
	if pkgInfo.Link != "" {
		return pkgInfo.Link
	}
	return "https://github.com/" + pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
}

func (pkgInfo *GitHubContentPackageInfo) GetDescription() string {
	return pkgInfo.Description
}

func (pkgInfo *GitHubContentPackageInfo) GetFormat() string {
	for _, arcTypeOverride := range pkgInfo.FormatOverrides {
		if arcTypeOverride.GOOS == runtime.GOOS {
			return arcTypeOverride.Format
		}
	}
	return pkgInfo.Format
}

func (pkgInfo *GitHubContentPackageInfo) GetReplacements() map[string]string {
	return pkgInfo.Replacements
}

func (pkgInfo *GitHubContentPackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
}

func (pkgInfo *GitHubContentPackageInfo) GetFiles() []*File {
	if len(pkgInfo.Files) != 0 {
		return pkgInfo.Files
	}
	return []*File{
		{
			Name: pkgInfo.RepoName,
		},
	}
}

func (pkgInfo *GitHubContentPackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
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

func (pkgInfo *GitHubContentPackageInfo) RenderAsset(pkg *Package) (string, error) {
	return pkgInfo.Path.Execute(map[string]interface{}{
		"Version": pkg.Version,
		"GOOS":    runtime.GOOS,
		"GOARCH":  runtime.GOARCH,
		"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":    replace(runtime.GOARCH, pkgInfo.GetReplacements()),
		"Format":  pkgInfo.GetFormat(),
	})
}
