package controller

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"

	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
)

type HTTPPackageInfo struct {
	Name            string `validate:"required"`
	Format          string
	Description     string
	Link            string
	Files           []*File `validate:"required,dive"`
	Replacements    map[string]string
	FormatOverrides []*FormatOverride

	URL *text.Template `validate:"required"`
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
