package config

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

// RenameFile renames files with appropriate Windows extensions when necessary.
// It handles Windows-specific file extension requirements for executable files.
func (p *Package) RenameFile(logger *slog.Logger, fs afero.Fs, pkgPath string, file *registry.File, rt *runtime.Runtime) (string, error) {
	s, err := p.fileSrcWithoutWindowsExt(file, rt)
	if err != nil {
		return "", err
	}
	if !rt.IsWindows() {
		return s, nil
	}
	if osfile.Ext(s, p.Package.Version) != "" {
		return s, nil
	}

	return p.renameFile(logger, fs, pkgPath, s)
}

// renameFile performs the actual file renaming with Windows extension.
// It checks if the target file already exists before attempting to rename.
func (p *Package) renameFile(logger *slog.Logger, fs afero.Fs, pkgPath, oldName string) (string, error) {
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
	logger.Info("rename a file", "new", newPath, "old", old)
	if err := fs.Rename(old, newPath); err != nil {
		return "", fmt.Errorf("rename a file: %w", err)
	}
	return newName, nil
}

// windowsExt returns the appropriate Windows file extension for the package.
// It uses package-specific configuration or defaults based on package type.
func (p *Package) windowsExt() string {
	if p.PackageInfo.WindowsExt == "" {
		if p.PackageInfo.Type == registry.PkgInfoTypeGitHubContent || p.PackageInfo.Type == registry.PkgInfoTypeGitHubArchive {
			return ".sh"
		}
		return ".exe"
	}
	return p.PackageInfo.WindowsExt
}

// completeWindowsExt adds Windows extension to a string if needed.
// It respects package configuration for extension completion behavior.
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

// completeWindowsExtToAsset adds Windows extension to asset names when appropriate.
// It considers file format and existing extensions to determine if completion is needed.
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
	if osfile.Ext(asset, p.Package.Version) == "" {
		return p.completeWindowsExt(asset)
	}
	return asset
}

// completeWindowsExtToURL adds Windows extension to URLs when appropriate.
// It delegates to asset extension logic for consistent behavior.
func (p *Package) completeWindowsExtToURL(url string) string {
	return p.completeWindowsExtToAsset(url)
}

// completeWindowsExtToFileSrc adds Windows extension to file source paths.
// It checks for existing extensions before adding the Windows extension.
func (p *Package) completeWindowsExtToFileSrc(src string) string {
	if osfile.Ext(src, p.Package.Version) == "" {
		return src + p.windowsExt()
	}
	return src
}
