package config

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
)

type Package struct {
	Package     *aqua.Package
	PackageInfo *registry.PackageInfo
	Type        PackageType
}

type PackageType interface {
	Install(ctx context.Context, pkg *Package, logE *logrus.Entry) error
	CheckInstalled(pkg *Package) (bool, error)
	GetFiles(pkgInfo *registry.PackageInfo) []*registry.File
	Find(pkg *Package, exeName string, logE *logrus.Entry) (string, error)
	GetFilePath(pkg *Package, file *registry.File) (string, error)
	GetFormat(pkgInfo *registry.PackageInfo) string
	GetName(pkgInfo *registry.PackageInfo) string
}

type PackageTypes = map[string]PackageType

func (cpkg *Package) RenderSrc(file *registry.File, rt *runtime.Runtime) (string, error) {
	pkg := cpkg.Package
	pkgInfo := cpkg.PackageInfo
	return template.Execute(file.Src, map[string]interface{}{ //nolint:wrapcheck
		"Version":  pkg.Version,
		"GOOS":     rt.GOOS,
		"GOARCH":   rt.GOARCH,
		"OS":       replace(rt.GOOS, pkgInfo.Replacements),
		"Arch":     getArch(pkgInfo.GetRosetta2(), pkgInfo.Replacements, rt),
		"Format":   cpkg.Type.GetFormat(pkgInfo),
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
