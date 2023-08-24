package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (p *Package) RenameFile(logE *logrus.Entry, fs afero.Fs, pkgPath string, file *registry.File, rt *runtime.Runtime) (string, error) {
	s, err := p.getFileSrcWithoutWindowsExt(file, rt)
	if err != nil {
		return "", err
	}
	if !isWindows(rt.GOOS) {
		return s, nil
	}
	if util.Ext(s, p.Package.Version) != "" {
		return s, nil
	}

	return p.renameFile(logE, fs, pkgPath, s)
}

func (p *Package) renameFile(logE *logrus.Entry, fs afero.Fs, pkgPath, oldName string) (string, error) {
	newName := oldName + p.windowsExt()
	newPath := filepath.Join(pkgPath, newName)
	if _, err := fs.Stat(newPath); err == nil {
		return newName, nil
	}
	old := filepath.Join(pkgPath, oldName)
	if _, err := fs.Stat(old); err != nil {
		return "", &FileNotFoundError{
			Err: err,
		}
	}
	logE.WithFields(logrus.Fields{
		"new": newPath,
		"old": old,
	}).Info("rename a file")
	if err := fs.Rename(old, newPath); err != nil {
		return "", fmt.Errorf("rename a file: %w", err)
	}
	return newName, nil
}

func (p *Package) windowsExt() string {
	if p.PackageInfo.WindowsExt == "" {
		if p.PackageInfo.Type == registry.PkgInfoTypeGitHubContent || p.PackageInfo.Type == registry.PkgInfoTypeGitHubArchive {
			return ".sh"
		}
		return ".exe"
	}
	return p.PackageInfo.WindowsExt
}

func (p *Package) completeWindowsExt(s string) string {
	if p.PackageInfo.CompleteWindowsExt != nil {
		if *p.PackageInfo.CompleteWindowsExt {
			return s + p.windowsExt()
		}
		return s
	}
	if p.PackageInfo.Type == registry.PkgInfoTypeGitHubContent || p.PackageInfo.Type == registry.PkgInfoTypeGitHubArchive {
		return s
	}
	return s + p.windowsExt()
}

func (p *Package) completeWindowsExtToAsset(asset string) string {
	if strings.HasSuffix(asset, ".exe") {
		return asset
	}
	if p.PackageInfo.Format == "raw" {
		return p.completeWindowsExt(asset)
	}
	if p.PackageInfo.Format != "" {
		return asset
	}
	if util.Ext(asset, p.Package.Version) == "" {
		return p.completeWindowsExt(asset)
	}
	return asset
}

func (p *Package) completeWindowsExtToURL(url string) string {
	return p.completeWindowsExtToAsset(url)
}

func (p *Package) completeWindowsExtToFileSrc(src string) string {
	if util.Ext(src, p.Package.Version) == "" {
		return src + p.windowsExt()
	}
	return src
}
