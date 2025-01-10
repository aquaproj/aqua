package config

import (
	"errors"
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
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
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

	if p.PackageInfo.GetAppendExt() {
		asset = appendExt(asset, p.PackageInfo.Format)
	}

	if !rt.IsWindows() {
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
		Arch:    getArch(pkgInfo.Rosetta2, pkgInfo.WindowsARMEmulation, pkgInfo.Replacements, rt),
		Format:  pkgInfo.GetFormat(),
		Asset:   asset,
		Vars:    pkg.Vars,
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
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, registry, pkgInfo.Crate, strings.TrimPrefix(pkg.Version, "v")), nil
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

	if p.PackageInfo.GetAppendExt() {
		s = appendExt(s, p.PackageInfo.Format)
	}

	if !rt.IsWindows() {
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
		"Arch":            getArch(pkgInfo.Rosetta2, pkgInfo.WindowsARMEmulation, pkgInfo.Replacements, rt),
		"Format":          format,
		"FileName":        file.Name,
		"Asset":           assetName,
		"AssetWithoutExt": assetWithoutExt,
		"Vars":            pkg.Vars,
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

func getArch(rosetta2, windowsARMEmulation bool, replacements registry.Replacements, rt *runtime.Runtime) string {
	if rosetta2 && rt.GOOS == "darwin" && rt.GOARCH == "arm64" {
		// Rosetta 2
		return replace("amd64", replacements)
	}
	if windowsARMEmulation && rt.GOOS == "windows" && rt.GOARCH == "arm64" {
		// Windows ARM Emulation
		return replace("amd64", replacements)
	}
	return replace(rt.GOARCH, replacements)
}

func (p *Package) fileSrc(file *registry.File, rt *runtime.Runtime) (string, error) {
	s, err := p.fileSrcWithoutWindowsExt(file, rt)
	if err != nil {
		return "", err
	}
	if !rt.IsWindows() {
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

type RemoveMode struct {
	Link    bool
	Package bool
}

type Param struct {
	GlobalConfigFilePaths             []string
	ConfigFilePath                    string
	LogLevel                          string
	File                              string
	AQUAVersion                       string
	AquaCommitHash                    string
	RootDir                           string
	PWD                               string
	InsertFile                        string
	LogColor                          string
	Dest                              string
	HomeDir                           string
	OutTestData                       string
	Limit                             int
	MaxParallelism                    int
	Args                              []string
	Tags                              map[string]struct{}
	ExcludedTags                      map[string]struct{}
	DisableLazyInstall                bool
	OnlyLink                          bool
	All                               bool
	Global                            bool
	Insert                            bool
	SelectVersion                     bool
	ShowVersion                       bool
	ProgressBar                       bool
	Deep                              bool
	SkipLink                          bool
	Pin                               bool
	Prune                             bool
	Checksum                          bool
	RequireChecksum                   bool
	EnforceChecksum                   bool
	EnforceRequireChecksum            bool
	DisablePolicy                     bool
	Detail                            bool
	OnlyPackage                       bool
	OnlyRegistry                      bool
	CosignDisabled                    bool
	GitHubArtifactAttestationDisabled bool
	SLSADisabled                      bool
	Installed                         bool
	PolicyConfigFilePaths             []string
	Commands                          []string
	VacuumDays                        int
}

func appendExt(s, format string) string {
	if _, f := asset.RemoveExtFromAsset(s); f != "raw" {
		return s
	}
	if format == formatRaw || format == "" {
		return s
	}
	if strings.HasSuffix(s, "."+format) {
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
		return path.Base(pkgInfo.GetFiles()[0].Name), nil
	case PkgInfoTypeGitHubContent:
		s, err := p.RenderTemplateString(pkgInfo.Path, rt)
		if err != nil {
			return "", fmt.Errorf("render a package path: %w", err)
		}
		return s, nil
	case PkgInfoTypeGitHubRelease:
		return p.RenderTemplateString(pkgInfo.Asset, rt)
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
		"Arch":    getArch(pkgInfo.Rosetta2, pkgInfo.WindowsARMEmulation, replacements, rt),
		"Format":  pkgInfo.GetFormat(),
		"Asset":   asset,
		"Vars":    pkg.Vars,
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
		"Arch":    getArch(pkgInfo.Rosetta2, pkgInfo.WindowsARMEmulation, pkgInfo.Replacements, rt),
		"Format":  pkgInfo.GetFormat(),
		"Vars":    pkg.Vars,
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
		"Arch":     getArch(pkgInfo.Rosetta2, pkgInfo.WindowsARMEmulation, pkgInfo.Replacements, rt),
		"Format":   pkgInfo.GetFormat(),
		"FileName": file.Name,
		"Vars":     pkg.Vars,
	})
}

func (p *Package) ApplyVars() error {
	if p.PackageInfo.Vars == nil {
		return nil
	}
	for _, v := range p.PackageInfo.Vars {
		if v.Name == "" {
			return errors.New("a variable name is empty")
		}
		if err := p.applyVar(v); err != nil {
			return fmt.Errorf("apply a variable: %w", logerr.WithFields(err, logrus.Fields{
				"var_name": v.Name,
			}))
		}
	}
	return nil
}

func (p *Package) applyVar(v *registry.Var) error {
	if _, ok := p.Package.Vars[v.Name]; ok {
		return nil
	}
	if v.Default != nil {
		if p.Package.Vars == nil {
			p.Package.Vars = map[string]any{
				v.Name: v.Default,
			}
			return nil
		}
		p.Package.Vars[v.Name] = v.Default
		return nil
	}
	if !v.Required {
		return nil
	}
	return errors.New("a variable is required")
}
