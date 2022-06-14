package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/aquaproj/aqua/pkg/unarchive"
)

type Package struct {
	Package     *aqua.Package
	PackageInfo *registry.PackageInfo
}

func (cpkg *Package) RenderSrc(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkg := cpkg.Package
	pkgInfo := cpkg.PackageInfo
	return template.Execute(file.Src, map[string]interface{}{ //nolint:wrapcheck
		"Version":  pkg.Version,
		"GOOS":     rt.GOOS,
		"GOARCH":   rt.GOARCH,
		"OS":       replace(rt.GOOS, pkgInfo.GetReplacements()),
		"Arch":     getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements(), rt),
		"Format":   pkgInfo.GetFormat(),
		"FileName": file.Name,
	})
}

func replace(key string, replacements map[string]string) string {
	a := replacements[key]
	if a == "" {
		return key
	}
	return a
}

func getArch(rosetta2 bool, replacements map[string]string, rt *runtime.Runtime) string {
	if rosetta2 && rt.GOOS == "darwin" && rt.GOARCH == "arm64" {
		// Rosetta 2
		return replace("amd64", replacements)
	}
	return replace(rt.GOARCH, replacements)
}

func (cpkg *Package) RenderDir(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	return template.Execute(file.Dir, map[string]interface{}{ //nolint:wrapcheck
		"Version":  pkg.Version,
		"GOOS":     rt.GOOS,
		"GOARCH":   rt.GOARCH,
		"OS":       replace(rt.GOOS, pkgInfo.GetReplacements()),
		"Arch":     getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements(), rt),
		"Format":   pkgInfo.GetFormat(),
		"FileName": file.Name,
	})
}

func setWindowsExeExt(src string, rt *runtime.Runtime) string {
	if rt.GOOS != "windows" || strings.HasSuffix(src, ".exe") {
		return src
	}
	return src + ".exe"
}

func (cpkg *Package) GetFileSrc(file *registry.File, rt *runtime.Runtime) (string, error) {
	s, err := cpkg.getFileSrc(file, rt)
	if err != nil {
		return "", err
	}
	return setWindowsExeExt(s, rt), nil
}

func (cpkg *Package) getFileSrc(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
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
	src, err := cpkg.RenderSrc(file, rt)
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
	PkgInfoTypeGo            = "go"
	PkgInfoTypeGoInstall     = "go_install"
)

type Param struct {
	ConfigFilePath        string
	LogLevel              string
	OnlyLink              bool
	IsTest                bool
	All                   bool
	Insert                bool
	File                  string
	GlobalConfigFilePaths []string
	AQUAVersion           string
	RootDir               string
	MaxParallelism        int
	PWD                   string
}

func (cpkg *Package) RenderAsset(rt *runtime.Runtime) (string, error) {
	asset, err := cpkg.renderAsset(rt)
	if err != nil {
		return "", err
	}
	if rt.GOOS == "windows" && !strings.HasSuffix(asset, ".exe") {
		if cpkg.PackageInfo.Format == "raw" {
			return asset + ".exe", nil
		}
		if cpkg.PackageInfo.Format != "" {
			return asset, nil
		}
		if filepath.Ext(asset) == "" {
			return asset + ".exe", nil
		}
	}
	return asset, nil
}

func (cpkg *Package) renderAsset(rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGo:
		return "", nil
	case PkgInfoTypeGoInstall:
		if pkgInfo.Asset != nil {
			return *pkgInfo.Asset, nil
		}
		return filepath.Base(pkgInfo.GetPath()), nil
	case PkgInfoTypeGitHubContent:
		s, err := cpkg.renderTemplateString(*pkgInfo.Path, rt)
		if err != nil {
			return "", fmt.Errorf("render a package path: %w", err)
		}
		return s, nil
	case PkgInfoTypeGitHubRelease:
		return cpkg.renderTemplateString(*pkgInfo.Asset, rt)
	case PkgInfoTypeHTTP:
		uS, err := cpkg.RenderURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return filepath.Base(u.Path), nil
	}
	return "", nil
}

func (cpkg *Package) renderTemplateString(s string, rt *runtime.Runtime) (string, error) {
	tpl, err := template.Compile(s)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	return cpkg.renderTemplate(tpl, rt)
}

func (cpkg *Package) renderTemplate(tpl *texttemplate.Template, rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	uS, err := template.ExecuteTemplate(tpl, map[string]interface{}{
		"Version": pkg.Version,
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

func (cpkg *Package) RenderURL(rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	return cpkg.renderTemplateString(*pkgInfo.URL, rt)
}

func (cpkg *Package) GetPkgPath(rootDir string, rt *runtime.Runtime) (string, error) {
	pkgInfo := cpkg.PackageInfo
	pkg := cpkg.Package
	assetName, err := cpkg.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGo:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, "src"), nil
	case PkgInfoTypeGoInstall:
		return filepath.Join(rootDir, "pkgs", pkgInfo.GetType(), pkgInfo.GetPath(), pkg.Version, "bin"), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
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
