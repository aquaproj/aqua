package config

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Package struct {
	Package     *aqua.Package
	PackageInfo *registry.PackageInfo
	Registry    *aqua.Registry
}

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

func (cpkg *Package) GetExePath(rootDir string, file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgPath, err := cpkg.GetPkgPath(rootDir, rt)
	if err != nil {
		return "", err
	}
	fileSrc, err := cpkg.getFileSrc(file, rt)
	if err != nil {
		return "", fmt.Errorf("get a file path: %w", err)
	}
	return filepath.Join(pkgPath, fileSrc), nil
}

func (cpkg *Package) RenderAsset(rt *runtime.Runtime) (string, error) {
	asset, err := cpkg.renderAsset(rt)
	if err != nil {
		return "", err
	}
	if asset == "" {
		return "", nil
	}
	if isWindows(rt.GOOS) && !strings.HasSuffix(asset, ".exe") {
		if cpkg.PackageInfo.Format == "raw" {
			return cpkg.completeWindowsExt(asset), nil
		}
		if cpkg.PackageInfo.Format != "" {
			return asset, nil
		}
		if util.Ext(asset, cpkg.Package.Version) == "" {
			return cpkg.completeWindowsExt(asset), nil
		}
	}
	return asset, nil
}

func (cpkg *Package) GetTemplateArtifact(rt *runtime.Runtime, asset string) *template.Artifact {
	pkg := cpkg.Package
	pkgInfo := cpkg.PackageInfo
	return &template.Artifact{
		Version: pkg.Version,
		SemVer:  cpkg.semVer(),
		OS:      replace(rt.GOOS, pkgInfo.GetReplacements()),
		Arch:    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements(), rt),
		Format:  pkgInfo.GetFormat(),
		Asset:   asset,
	}
}

func (cpkg *Package) RenderPath() (string, error) {
	pkgInfo := cpkg.PackageInfo
	return cpkg.RenderTemplateString(pkgInfo.GetPath(), &runtime.Runtime{})
}

func (cpkg *Package) GetPkgPath(rootDir string, rt *runtime.Runtime) (string, error) { //nolint:cyclop
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	assetName, err := cpkg.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGoInstall:
		p, err := cpkg.RenderPath()
		if err != nil {
			return "", fmt.Errorf("render Go Module Path: %w", err)
		}
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), p, pkg.Version, "bin"), nil
	case PkgInfoTypeCargo:
		registry := "crates.io"
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), registry, *pkgInfo.Crate, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		if pkgInfo.RepoOwner == "aquaproj" && (pkgInfo.RepoName == "aqua" || pkgInfo.RepoName == "aqua-proxy") {
			return filepath.Join(rootDir, "internal", "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
		}
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
		uS, err := cpkg.RenderURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), u.Host, u.Path), nil
	}
	return "", nil
}

func (cpkg *Package) RenderTemplateString(s string, rt *runtime.Runtime) (string, error) {
	tpl, err := template.Compile(s)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	return cpkg.renderTemplate(tpl, rt)
}

func (cpkg *Package) RenderURL(rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	s, err := cpkg.RenderTemplateString(*pkgInfo.URL, rt)
	if err != nil {
		return "", err
	}
	if isWindows(rt.GOOS) && !strings.HasSuffix(s, ".exe") {
		if cpkg.PackageInfo.Format == "raw" {
			return cpkg.completeWindowsExt(s), nil
		}
		if cpkg.PackageInfo.Format != "" {
			return s, nil
		}
		if util.Ext(s, cpkg.Package.Version) == "" {
			return cpkg.completeWindowsExt(s), nil
		}
	}
	return s, nil
}

type FileNotFoundError struct {
	Err error
}

func (errorFileNotFound *FileNotFoundError) Error() string {
	return errorFileNotFound.Err.Error()
}

func (errorFileNotFound *FileNotFoundError) Unwrap() error {
	return errorFileNotFound.Err
}

func isWindows(goos string) bool {
	return goos == "windows"
}

func (cpkg *Package) renderSrc(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkg := cpkg.Package
	pkgInfo := cpkg.PackageInfo
	s, err := template.Execute(file.Src, map[string]interface{}{
		"Version":  pkg.Version,
		"SemVer":   cpkg.semVer(),
		"GOOS":     rt.GOOS,
		"GOARCH":   rt.GOARCH,
		"OS":       replace(rt.GOOS, pkgInfo.GetReplacements()),
		"Arch":     getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements(), rt),
		"Format":   pkgInfo.GetFormat(),
		"FileName": file.Name,
	})
	if err != nil {
		return "", err //nolint:wrapcheck
	}
	return filepath.FromSlash(s), nil // FromSlash is needed for Windows. https://github.com/aquaproj/aqua/issues/2013
}

func replace(key string, replacements registry.Replacements) string {
	a := replacements[key]
	if a == "" {
		return key
	}
	return a
}

func getArch(rosetta2 bool, replacements registry.Replacements, rt *runtime.Runtime) string {
	if rosetta2 && rt.GOOS == "darwin" && rt.GOARCH == "arm64" {
		// Rosetta 2
		return replace("amd64", replacements)
	}
	return replace(rt.GOARCH, replacements)
}

func (cpkg *Package) getFileSrc(file *registry.File, rt *runtime.Runtime) (string, error) {
	s, err := cpkg.getFileSrcWithoutWindowsExt(file, rt)
	if err != nil {
		return "", err
	}
	if isWindows(rt.GOOS) && util.Ext(s, cpkg.Package.Version) == "" {
		return s + cpkg.windowsExt(), nil
	}
	return s, nil
}

func (cpkg *Package) getFileSrcWithoutWindowsExt(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	if pkgInfo.Type == "cargo" {
		return filepath.Join("bin", file.Name), nil
	}
	assetName, err := cpkg.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if unarchive.IsUnarchived(pkgInfo.GetFormat(), assetName) {
		return filepath.Base(assetName), nil
	}
	if file.Src == "" {
		return file.Name, nil
	}
	src, err := cpkg.renderSrc(file, rt)
	if err != nil {
		return "", fmt.Errorf("render the template file.src: %w", err)
	}
	return src, nil
}

const (
	PkgInfoTypeGitHubRelease = "github_release"
	PkgInfoTypeGitHubContent = "github_content"
	PkgInfoTypeGitHubArchive = "github_archive"
	PkgInfoTypeHTTP          = "http"
	PkgInfoTypeGoInstall     = "go_install"
	PkgInfoTypeCargo         = "cargo"
)

type Param struct {
	GlobalConfigFilePaths []string
	ConfigFilePath        string
	LogLevel              string
	File                  string
	AQUAVersion           string
	RootDir               string
	PWD                   string
	InsertFile            string
	LogColor              string
	Dest                  string
	HomeDir               string
	OutTestData           string
	MaxParallelism        int
	Args                  []string
	Tags                  map[string]struct{}
	ExcludedTags          map[string]struct{}
	OnlyLink              bool
	All                   bool
	Insert                bool
	SelectVersion         bool
	ProgressBar           bool
	Deep                  bool
	SkipLink              bool
	Pin                   bool
	Prune                 bool
	RequireChecksum       bool
	DisablePolicy         bool
	PolicyConfigFilePaths []string
}

func (cpkg *Package) renderAsset(rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return "", nil
	case PkgInfoTypeGoInstall:
		if pkgInfo.Asset != nil {
			return *pkgInfo.Asset, nil
		}
		return path.Base(pkgInfo.GetPath()), nil
	case PkgInfoTypeGitHubContent:
		s, err := cpkg.RenderTemplateString(*pkgInfo.Path, rt)
		if err != nil {
			return "", fmt.Errorf("render a package path: %w", err)
		}
		return s, nil
	case PkgInfoTypeGitHubRelease:
		return cpkg.RenderTemplateString(*pkgInfo.Asset, rt)
	case PkgInfoTypeHTTP:
		uS, err := cpkg.RenderURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return path.Base(u.Path), nil
	}
	return "", nil
}

func (cpkg *Package) renderChecksumFile(asset string, rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	tpl, err := template.Compile(pkgInfo.Checksum.Asset)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	replacements := pkgInfo.GetChecksumReplacements()
	uS, err := template.ExecuteTemplate(tpl, map[string]interface{}{
		"Version": pkg.Version,
		"SemVer":  cpkg.semVer(),
		"GOOS":    rt.GOOS,
		"GOARCH":  rt.GOARCH,
		"OS":      replace(rt.GOOS, replacements),
		"Arch":    getArch(pkgInfo.GetRosetta2(), replacements, rt),
		"Format":  pkgInfo.GetFormat(),
		"Asset":   asset,
	})
	if err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return uS, nil
}

func (cpkg *Package) renderTemplate(tpl *texttemplate.Template, rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	uS, err := template.ExecuteTemplate(tpl, map[string]interface{}{
		"Version": pkg.Version,
		"SemVer":  cpkg.semVer(),
		"GOOS":    rt.GOOS,
		"GOARCH":  rt.GOARCH,
		"OS":      replace(rt.GOOS, pkgInfo.GetReplacements()),
		"Arch":    getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements(), rt),
		"Format":  pkgInfo.GetFormat(),
	})
	if err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return uS, nil
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

func (cpkg *Package) semVer() string {
	v := cpkg.Package.Version
	prefix := cpkg.PackageInfo.GetVersionPrefix()
	if prefix == "" {
		return v
	}
	return strings.TrimPrefix(v, prefix)
}
