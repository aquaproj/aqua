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

func (cpkg *Package) RenameFile(logE *logrus.Entry, fs afero.Fs, pkgPath string, file *registry.File, rt *runtime.Runtime) (string, error) {
	s, err := cpkg.getFileSrc(file, rt)
	if err != nil {
		return "", err
	}
	if !(isWindows(rt.GOOS) && util.Ext(s, cpkg.Package.Version) == "") {
		return s, nil
	}
	newName := s + cpkg.windowsExt()
	newPath := filepath.Join(pkgPath, newName)
	if s == newName {
		return newName, nil
	}
	if _, err := fs.Stat(newPath); err == nil {
		return newName, nil
	}
	old := filepath.Join(pkgPath, s)
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

func (cpkg *Package) windowsExt() string {
	if cpkg.PackageInfo.WindowsExt == "" {
		if cpkg.PackageInfo.Type == registry.PkgInfoTypeGitHubContent || cpkg.PackageInfo.Type == registry.PkgInfoTypeGitHubArchive {
			return ".sh"
		}
		return ".exe"
	}
	return cpkg.PackageInfo.WindowsExt
}

func (cpkg *Package) completeWindowsExt(s string) string {
	if cpkg.PackageInfo.CompleteWindowsExt != nil {
		if *cpkg.PackageInfo.CompleteWindowsExt {
			return s + cpkg.windowsExt()
		}
		return s
	}
	if cpkg.PackageInfo.Type == registry.PkgInfoTypeGitHubContent || cpkg.PackageInfo.Type == registry.PkgInfoTypeGitHubArchive {
		return s
	}
	return s + cpkg.windowsExt()
}

func (cpkg *Package) completeWindowsExtToAsset(asset string) string {
	if strings.HasSuffix(asset, ".exe") {
		return asset
	}
	if cpkg.PackageInfo.Format == "raw" {
		return cpkg.completeWindowsExt(asset)
	}
	if cpkg.PackageInfo.Format != "" {
		return asset
	}
	if util.Ext(asset, cpkg.Package.Version) == "" {
		return cpkg.completeWindowsExt(asset)
	}
	return asset
}

func (cpkg *Package) completeWindowsExtToURL(url string) string {
	if strings.HasSuffix(url, ".exe") {
		return url
	}
	if cpkg.PackageInfo.Format == "raw" {
		return cpkg.completeWindowsExt(url)
	}
	if cpkg.PackageInfo.Format != "" {
		return url
	}
	if util.Ext(url, cpkg.Package.Version) == "" {
		return cpkg.completeWindowsExt(url)
	}
	return url
}

func (cpkg *Package) completeWindowsExtToFileSrc(src string) string {
	if util.Ext(src, cpkg.Package.Version) == "" {
		return src + cpkg.windowsExt()
	}
	return src
}
