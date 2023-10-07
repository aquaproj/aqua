package config

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
)

type Package struct {
	Package     *aqua.Package
	PackageInfo *registry.PackageInfo
	Registry    *aqua.Registry
}

func (p *Package) ExePath(rootDir string, file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	if pkgInfo.Type == "go_build" {
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, p.Package.Version, "bin", file.Name), nil
	}

	pkgPath, err := p.PkgPath(rootDir, rt)
	if err != nil {
		return "", err
	}
	fileSrc, err := p.fileSrc(file, rt)
	if err != nil {
		return "", fmt.Errorf("get a file path: %w", err)
	}
	return filepath.Join(pkgPath, fileSrc), nil
}

func (p *Package) RenderAsset(rt *runtime.Runtime) (string, error) {
	asset, err := p.renderAsset(rt)
	if err != nil {
		return "", err
	}
	if asset == "" {
		return "", nil
	}

	if p.PackageInfo.GetAppendFormat() {
		asset = appendExt(asset, p.PackageInfo.Format)
	}

	if !isWindows(rt.GOOS) {
		return asset, nil
	}
	return p.completeWindowsExtToAsset(asset), nil
}

func (p *Package) TemplateArtifact(rt *runtime.Runtime, asset string) *template.Artifact {
	pkg := p.Package
	pkgInfo := p.PackageInfo
	return &template.Artifact{
		Version: pkg.Version,
		SemVer:  p.semVer(),
		OS:      replace(rt.GOOS, pkgInfo.Replacements),
		Arch:    getArch(pkgInfo.Rosetta2, pkgInfo.Replacements, rt),
		Format:  pkgInfo.GetFormat(),
		Asset:   asset,
	}
}

func (p *Package) RenderPath() (string, error) {
	pkgInfo := p.PackageInfo
	return p.RenderTemplateString(pkgInfo.GetPath(), &runtime.Runtime{})
}

func (p *Package) PkgPath(rootDir string, rt *runtime.Runtime) (string, error) { //nolint:cyclop
	pkgInfo := p.PackageInfo
	pkg := p.Package
	assetName, err := p.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGoBuild:
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, "src"), nil
	case PkgInfoTypeGoInstall:
		p, err := p.RenderPath()
		if err != nil {
			return "", fmt.Errorf("render Go Module Path: %w", err)
		}
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, p, pkg.Version, "bin"), nil
	case PkgInfoTypeCargo:
		registry := "crates.io"
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, registry, pkgInfo.Crate, pkg.Version), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		if pkgInfo.RepoOwner == "aquaproj" && (pkgInfo.RepoName == "aqua" || pkgInfo.RepoName == "aqua-proxy") {
			return filepath.Join(rootDir, "internal", "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
		}
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
		uS, err := p.RenderURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, u.Host, u.Path), nil
	}
	return "", nil
}

func (p *Package) RenderTemplateString(s string, rt *runtime.Runtime) (string, error) {
	tpl, err := template.Compile(s)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	return p.renderTemplate(tpl, rt)
}

func (p *Package) RenderURL(rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	s, err := p.RenderTemplateString(pkgInfo.URL, rt)
	if err != nil {
		return "", err
	}

	if p.PackageInfo.GetAppendFormat() {
		s = appendExt(s, p.PackageInfo.Format)
	}

	if !isWindows(rt.GOOS) {
		return s, nil
	}
	return p.completeWindowsExtToURL(s), nil
}

type FileNotFoundError struct {
	Err error
}

func (e *FileNotFoundError) Error() string {
	return e.Err.Error()
}

func (e *FileNotFoundError) Unwrap() error {
	return e.Err
}

func (p *Package) renderSrc(assetName string, file *registry.File, rt *runtime.Runtime) (string, error) {
	pkg := p.Package
	pkgInfo := p.PackageInfo
	format := pkgInfo.GetFormat()
	assetWithoutExt, _ := asset.RemoveExtFromAsset(assetName)
	s, err := template.Execute(file.Src, map[string]interface{}{
		"Version":         pkg.Version,
		"SemVer":          p.semVer(),
		"GOOS":            rt.GOOS,
		"GOARCH":          rt.GOARCH,
		"OS":              replace(rt.GOOS, pkgInfo.Replacements),
		"Arch":            getArch(pkgInfo.Rosetta2, pkgInfo.Replacements, rt),
		"Format":          format,
		"FileName":        file.Name,
		"Asset":           assetName,
		"AssetWithoutExt": assetWithoutExt,
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

func (p *Package) fileSrc(file *registry.File, rt *runtime.Runtime) (string, error) {
	s, err := p.fileSrcWithoutWindowsExt(file, rt)
	if err != nil {
		return "", err
	}
	if !isWindows(rt.GOOS) {
		return s, nil
	}
	return p.completeWindowsExtToFileSrc(s), nil
}

func (p *Package) fileSrcWithoutWindowsExt(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	if pkgInfo.Type == "cargo" {
		return filepath.Join("bin", file.Name), nil
	}
	assetName, err := p.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if unarchive.IsUnarchived(pkgInfo.GetFormat(), assetName) {
		return filepath.Base(assetName), nil
	}
	if file.Src == "" {
		return file.Name, nil
	}
	src, err := p.renderSrc(assetName, file, rt)
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
	PkgInfoTypeGoBuild       = "go_build"
	PkgInfoTypeCargo         = "cargo"
)

type Param struct {
	GlobalConfigFilePaths []string
	ConfigFilePath        string
	LogLevel              string
	File                  string
	AQUAVersion           string
	AquaCommitHash        string
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
	DisableLazyInstall    bool
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
	Detail                bool
	PolicyConfigFilePaths []string
}

func appendExt(s, format string) string {
	if _, f := asset.RemoveExtFromAsset(s); f != "raw" {
		return s
	}
	if format == formatRaw || format == "" {
		return s
	}
	if strings.HasSuffix(s, fmt.Sprintf(".%s", format)) {
		return s
	}
	return fmt.Sprintf("%s.%s", s, format)
}

func (p *Package) renderAsset(rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive, PkgInfoTypeGoBuild:
		return "", nil
	case PkgInfoTypeGoInstall:
		if pkgInfo.Asset != nil {
			return *pkgInfo.Asset, nil
		}
		return path.Base(pkgInfo.GetPath()), nil
	case PkgInfoTypeGitHubContent:
		s, err := p.RenderTemplateString(pkgInfo.Path, rt)
		if err != nil {
			return "", fmt.Errorf("render a package path: %w", err)
		}
		return s, nil
	case PkgInfoTypeGitHubRelease:
		return p.RenderTemplateString(*pkgInfo.Asset, rt)
	case PkgInfoTypeHTTP:
		uS, err := p.RenderURL(rt)
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

func (p *Package) renderChecksumFile(asset string, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	pkg := p.Package
	tpl, err := template.Compile(pkgInfo.Checksum.Asset)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	replacements := pkgInfo.GetChecksumReplacements()
	uS, err := template.ExecuteTemplate(tpl, map[string]interface{}{
		"Version": pkg.Version,
		"SemVer":  p.semVer(),
		"GOOS":    rt.GOOS,
		"GOARCH":  rt.GOARCH,
		"OS":      replace(rt.GOOS, replacements),
		"Arch":    getArch(pkgInfo.Rosetta2, replacements, rt),
		"Format":  pkgInfo.GetFormat(),
		"Asset":   asset,
	})
	if err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return uS, nil
}

func (p *Package) renderTemplate(tpl *texttemplate.Template, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	pkg := p.Package
	uS, err := template.ExecuteTemplate(tpl, map[string]interface{}{
		"Version": pkg.Version,
		"SemVer":  p.semVer(),
		"GOOS":    rt.GOOS,
		"GOARCH":  rt.GOARCH,
		"OS":      replace(rt.GOOS, pkgInfo.Replacements),
		"Arch":    getArch(pkgInfo.Rosetta2, pkgInfo.Replacements, rt),
		"Format":  pkgInfo.GetFormat(),
	})
	if err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return uS, nil
}

func (p *Package) semVer() string {
	v := p.Package.Version
	prefix := p.PackageInfo.VersionPrefix
	if prefix == "" {
		return v
	}
	return strings.TrimPrefix(v, prefix)
}

func (p *Package) RenderDir(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	pkg := p.Package
	return template.Execute(file.Dir, map[string]interface{}{ //nolint:wrapcheck
		"Version":  pkg.Version,
		"SemVer":   p.semVer(),
		"GOOS":     rt.GOOS,
		"GOARCH":   rt.GOARCH,
		"OS":       replace(rt.GOOS, pkgInfo.Replacements),
		"Arch":     getArch(pkgInfo.Rosetta2, pkgInfo.Replacements, rt),
		"Format":   pkgInfo.GetFormat(),
		"FileName": file.Name,
	})
}
