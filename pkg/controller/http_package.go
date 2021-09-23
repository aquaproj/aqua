package controller

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
)

type HTTPPackageInfo struct {
	Name               string `validate:"required"`
	Format             string
	Description        string
	Link               string
	Files              []*File `validate:"required,dive"`
	Replacements       map[string]string
	FormatOverrides    []*FormatOverride
	VersionConstraints *VersionConstraints
	VersionOverrides   []*HTTPVersionOverride

	URL *Template `validate:"required"`
}

func (pkgInfo *HTTPPackageInfo) SetVersion(v string) (PackageInfo, error) {
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
			return overrideHTTPPackageInfo(pkgInfo, vo), nil
		}
	}
	return pkgInfo, nil
}

type HTTPVersionOverride struct {
	VersionConstraints *VersionConstraints
	URL                *Template `validate:"required"`
	Files              []*File   `validate:"dive"`
	Format             string
	FormatOverrides    []*FormatOverride
	Replacements       map[string]string
}

func overrideHTTPPackageInfo(base *HTTPPackageInfo, vo *HTTPVersionOverride) *HTTPPackageInfo {
	p := &HTTPPackageInfo{
		Name:            base.Name,
		Format:          base.Format,
		Link:            base.Link,
		Description:     base.Description,
		Files:           base.Files,
		Replacements:    base.Replacements,
		FormatOverrides: base.FormatOverrides,
		URL:             base.URL,
	}
	if vo.URL != nil {
		p.URL = vo.URL
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

func (pkgInfo *HTTPPackageInfo) GetName() string {
	return pkgInfo.Name
}

func (pkgInfo *HTTPPackageInfo) GetType() string {
	return pkgInfoTypeHTTP
}

func (pkgInfo *HTTPPackageInfo) GetLink() string {
	return pkgInfo.Link
}

func (pkgInfo *HTTPPackageInfo) GetDescription() string {
	return pkgInfo.Description
}

func (pkgInfo *HTTPPackageInfo) GetFormat() string {
	for _, arcTypeOverride := range pkgInfo.FormatOverrides {
		if arcTypeOverride.GOOS == runtime.GOOS {
			return arcTypeOverride.Format
		}
	}
	return pkgInfo.Format
}

func (pkgInfo *HTTPPackageInfo) GetReplacements() map[string]string {
	return pkgInfo.Replacements
}

func (pkgInfo *HTTPPackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	uS, err := pkgInfo.URL.Execute(map[string]interface{}{
		"Version": pkg.Version,
		"GOOS":    runtime.GOOS,
		"GOARCH":  runtime.GOARCH,
		"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":    replace(runtime.GOARCH, pkgInfo.GetReplacements()),
		"Format":  pkgInfo.GetFormat(),
	})
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}
	u, err := url.Parse(uS)
	if err != nil {
		return "", fmt.Errorf("parse the URL: %w", err)
	}
	return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), u.Host, u.Path), nil
}

func (pkgInfo *HTTPPackageInfo) GetFiles() []*File {
	return pkgInfo.Files
}

func (pkgInfo *HTTPPackageInfo) GetFileSrc(pkg *Package, file *File) (string, error) {
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

func (pkgInfo *HTTPPackageInfo) RenderURL(pkg *Package) (string, error) {
	uS, err := pkgInfo.URL.Execute(map[string]interface{}{
		"Version": pkg.Version,
		"GOOS":    runtime.GOOS,
		"GOARCH":  runtime.GOARCH,
		"OS":      replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":    replace(runtime.GOARCH, pkgInfo.GetReplacements()),
		"Format":  pkgInfo.GetFormat(),
	})
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}
	return uS, nil
}

func (pkgInfo *HTTPPackageInfo) RenderAsset(pkg *Package) (string, error) {
	uS, err := pkgInfo.RenderURL(pkg)
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}
	u, err := url.Parse(uS)
	if err != nil {
		return "", fmt.Errorf("parse the URL: %w", err)
	}
	return filepath.Base(u.Path), nil
}
