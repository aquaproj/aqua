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

// Package represents a complete package configuration with its metadata.
// It combines the package definition from aqua.yaml with registry information and registry details.
type Package struct {
	Package     *aqua.Package         // Package configuration from aqua.yaml
	PackageInfo *registry.PackageInfo // Package metadata from registry
	Registry    *aqua.Registry        // Registry information where package is defined
}

// ExePath returns the absolute path to an executable file for the package.
// It handles different package types and constructs the appropriate file system path.
func (p *Package) ExePath(rootDir string, file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	if pkgInfo.Type == "go_build" {
		return filepath.Join(rootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, p.Package.Version, "bin", file.Name), nil
	}

	pkgPath, err := p.AbsPkgPath(rootDir, rt)
	if err != nil {
		return "", err
	}
	fileSrc, err := p.fileSrc(file, rt)
	if err != nil {
		return "", fmt.Errorf("get a file path: %w", err)
	}
	return filepath.Join(pkgPath, fileSrc), nil
}

// RenderAsset renders the asset name for the given runtime.
// It applies templating, extensions, and platform-specific modifications.
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

// TemplateArtifact creates a template artifact with all necessary context variables.
// It provides template variables for rendering asset names, URLs, and paths.
func (p *Package) TemplateArtifact(rt *runtime.Runtime, aset string) *template.Artifact {
	pkg := p.Package
	pkgInfo := p.PackageInfo
	assetWithoutExt, _ := asset.RemoveExtFromAsset(aset)
	return &template.Artifact{
		Version:         pkg.Version,
		SemVer:          p.semVer(),
		OS:              replace(rt.GOOS, pkgInfo.Replacements),
		Arch:            getArch(pkgInfo.Rosetta2, pkgInfo.WindowsARMEmulation, pkgInfo.Replacements, rt),
		Format:          pkgInfo.GetFormat(),
		Asset:           aset,
		AssetWithoutExt: assetWithoutExt,
		Vars:            pkg.Vars,
	}
}

// RenderPath renders the package path using templates.
// Used primarily for Go modules and other path-based package types.
func (p *Package) RenderPath() (string, error) {
	pkgInfo := p.PackageInfo
	return p.RenderTemplateString(pkgInfo.GetPath(), &runtime.Runtime{})
}

// PkgPath returns the relative path where the package should be installed.
// The path format varies by package type (GitHub, HTTP, Go, Cargo, etc.).
func (p *Package) PkgPath(rt *runtime.Runtime) (string, error) { //nolint:cyclop
	pkgInfo := p.PackageInfo
	pkg := p.Package
	assetName, err := p.RenderAsset(rt)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	switch pkgInfo.Type {
	case PkgInfoTypeGitHubArchive:
		return filepath.Join("pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version), nil
	case PkgInfoTypeGoBuild:
		return filepath.Join("pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, "src"), nil
	case PkgInfoTypeGoInstall:
		p, err := p.RenderPath()
		if err != nil {
			return "", fmt.Errorf("render Go Module Path: %w", err)
		}
		return filepath.Join("pkgs", pkgInfo.Type, p, pkg.Version, "bin"), nil
	case PkgInfoTypeCargo:
		registry := "crates.io"
		return filepath.Join("pkgs", pkgInfo.Type, registry, pkgInfo.Crate, strings.TrimPrefix(pkg.Version, "v")), nil
	case PkgInfoTypeGitHubContent, PkgInfoTypeGitHubRelease:
		if pkgInfo.RepoOwner == "aquaproj" && (pkgInfo.RepoName == "aqua" || pkgInfo.RepoName == "aqua-proxy") {
			return filepath.Join("internal", "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
		}
		return filepath.Join("pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName), nil
	case PkgInfoTypeHTTP:
		uS, err := p.RenderURL(rt)
		if err != nil {
			return "", fmt.Errorf("render URL: %w", err)
		}
		u, err := url.Parse(uS)
		if err != nil {
			return "", fmt.Errorf("parse the URL: %w", err)
		}
		return filepath.Join("pkgs", pkgInfo.Type, u.Host, u.Path), nil
	}
	return "", nil
}

// AbsPkgPath returns the absolute path where the package should be installed.
// It combines the root directory with the package-specific path.
func (p *Package) AbsPkgPath(rootDir string, rt *runtime.Runtime) (string, error) {
	pkgPath, err := p.PkgPath(rt)
	if err != nil {
		return "", err
	}
	return filepath.Join(rootDir, pkgPath), nil
}

// RenderTemplateString renders a template string with package and runtime context.
// It compiles the template and executes it with all available template variables.
func (p *Package) RenderTemplateString(s string, rt *runtime.Runtime) (string, error) {
	tpl, err := template.Compile(s)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	return p.renderTemplate(tpl, rt)
}

// RenderURL renders the download URL for the package using templates.
// It handles platform-specific URL generation and extension handling.
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

// RenderDir renders the directory path for a file using templates.
// It provides template variables for platform-specific directory generation.
func (p *Package) RenderDir(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	pkg := p.Package
	return template.Execute(file.Dir, map[string]any{ //nolint:wrapcheck
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

// ApplyVars applies default values for package variables.
// It processes variable definitions and sets defaults for required variables.
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

// FileNotFoundError represents an error when a required file is not found.
// It wraps the underlying error while maintaining the error chain.
type FileNotFoundError struct {
	Err error
}

// Error returns the error message for FileNotFoundError.
func (e *FileNotFoundError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error for error chain unwrapping.
func (e *FileNotFoundError) Unwrap() error {
	return e.Err
}

// renderSrc renders the source path template for a file within a package.
// It provides template variables for generating platform-specific file paths.
func (p *Package) renderSrc(assetName string, file *registry.File, rt *runtime.Runtime) (string, error) {
	pkg := p.Package
	pkgInfo := p.PackageInfo
	format := pkgInfo.GetFormat()
	assetWithoutExt, _ := asset.RemoveExtFromAsset(assetName)
	s, err := template.Execute(file.Src, map[string]any{
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

// replace applies string replacements from the replacements map.
// If no replacement is found, it returns the original key.
func replace(key string, replacements registry.Replacements) string {
	a := replacements[key]
	if a == "" {
		return key
	}
	return a
}

// getArch returns the appropriate architecture string considering emulation.
// It handles Rosetta 2 on macOS ARM64 and Windows ARM emulation scenarios.
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

// fileSrc returns the source path for a file, including Windows extension handling.
// It determines the appropriate file path within the package structure.
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

// fileSrcWithoutWindowsExt returns the source path without Windows-specific extensions.
// It handles different package types and archive formats to determine the correct file path.
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

// Package type constants defining how packages are distributed and installed.
const (
	// PkgInfoTypeGitHubRelease indicates packages distributed via GitHub releases
	PkgInfoTypeGitHubRelease = "github_release"
	// PkgInfoTypeGitHubContent indicates packages downloaded from GitHub repository content
	PkgInfoTypeGitHubContent = "github_content"
	// PkgInfoTypeGitHubArchive indicates packages using GitHub's archive download
	PkgInfoTypeGitHubArchive = "github_archive"
	// PkgInfoTypeHTTP indicates packages downloaded from arbitrary HTTP URLs
	PkgInfoTypeHTTP = "http"
	// PkgInfoTypeGoInstall indicates packages installed via 'go install' command
	PkgInfoTypeGoInstall = "go_install"
	// PkgInfoTypeGoBuild indicates packages built from Go source code
	PkgInfoTypeGoBuild = "go_build"
	// PkgInfoTypeCargo indicates packages installed via Cargo (Rust package manager)
	PkgInfoTypeCargo = "cargo"
)

// RemoveMode specifies what should be removed during package removal operations.
type RemoveMode struct {
	Link    bool // Whether to remove symlinks
	Package bool // Whether to remove the package itself
}

// Param contains all configuration parameters and flags for aqua operations.
// It consolidates command-line flags, environment variables, and configuration settings.
type Param struct {
	ConfigFilePath                    string
	GenerateConfigFilePath            string
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
	AssetFile                         string
	Limit                             int
	MaxParallelism                    int
	VacuumDays                        int
	GlobalConfigFilePaths             []string
	Args                              []string
	PolicyConfigFilePaths             []string
	Commands                          []string
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
	GitHubReleaseAttestationDisabled  bool
	SLSADisabled                      bool
	Installed                         bool
	InitConfig                        bool
}

// appendExt appends the appropriate file extension based on format.
// It only adds extensions for raw format files that don't already have the extension.
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

// renderAsset renders the asset name without extensions or Windows-specific modifications.
// It handles different package types to generate the appropriate asset identifier.
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

// renderChecksumFile renders the checksum file name using templates.
// It provides template variables for generating platform-specific checksum file names.
func (p *Package) renderChecksumFile(asset string, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	pkg := p.Package
	tpl, err := template.Compile(pkgInfo.Checksum.Asset)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	replacements := pkgInfo.GetChecksumReplacements()
	uS, err := template.ExecuteTemplate(tpl, map[string]any{
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

// renderTemplate executes a template with package and runtime context.
// It provides all standard template variables for package rendering.
func (p *Package) renderTemplate(tpl *texttemplate.Template, rt *runtime.Runtime) (string, error) {
	pkgInfo := p.PackageInfo
	pkg := p.Package
	uS, err := template.ExecuteTemplate(tpl, map[string]any{
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

// semVer returns the semantic version by removing the version prefix.
// It strips configured prefixes (like 'v') to get clean semantic versions.
func (p *Package) semVer() string {
	v := p.Package.Version
	prefix := p.PackageInfo.VersionPrefix
	if prefix == "" {
		return v
	}
	return strings.TrimPrefix(v, prefix)
}

// applyVar applies a single variable definition to the package.
// It sets default values for variables that aren't already configured.
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
