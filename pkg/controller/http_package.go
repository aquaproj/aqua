package controller

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"

	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
)

type HTTPPackageInfo struct {
	Name        string  `validate:"required"`
	ArchiveType string  `yaml:"archive_type"`
	Files       []*File `validate:"required,dive"`

	URL *text.Template `validate:"required"`
}

func (pkgInfo *HTTPPackageInfo) GetName() string {
	return pkgInfo.Name
}

func (pkgInfo *HTTPPackageInfo) GetType() string {
	return pkgInfoTypeHTTP
}

func (pkgInfo *HTTPPackageInfo) GetArchiveType() string {
	return pkgInfo.ArchiveType
}

func (pkgInfo *HTTPPackageInfo) GetPkgPath(rootDir string, pkg *Package) (string, error) {
	uS, err := pkgInfo.URL.Execute(map[string]interface{}{
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
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

func (pkgInfo *HTTPPackageInfo) RenderURL(pkg *Package) (string, error) {
	uS, err := pkgInfo.URL.Execute(map[string]interface{}{
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
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
