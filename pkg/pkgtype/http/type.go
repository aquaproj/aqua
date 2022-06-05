package http

import (
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/apperr"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/render"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/spf13/afero"
)

const PkgType = "http"

func New(param *config.Param, fs afero.Fs, rt *runtime.Runtime) *Installer {
	return &Installer{
		fs:      fs,
		runtime: rt,
		rootDir: param.RootDir,
	}
}

type Installer struct {
	fs      afero.Fs
	runtime *runtime.Runtime
	http    download.HTTPDownloader
	rootDir string
}

func (inst *Installer) validate(pkgInfo *registry.PackageInfo) error {
	if pkgInfo.Name == "" {
		return apperr.ErrPkgNameIsRequired
	}
	if pkgInfo.URL == nil {
		return errURLRequired
	}
	return nil
}

func (inst *Installer) renderURL(pkg *config.Package) (string, error) {
	pkgInfo := pkg.PackageInfo
	rt := inst.runtime
	return template.Execute(*pkg.PackageInfo.URL, map[string]interface{}{ //nolint:wrapcheck
		"Version": pkg.Package.Version,
		"GOOS":    inst.runtime.GOOS,
		"GOARCH":  inst.runtime.GOARCH,
		"OS":      render.Replace(rt.GOOS, pkgInfo.Replacements),
		"Arch":    render.GetArch(pkgInfo.GetRosetta2(), pkgInfo.Replacements, rt),
		"Format":  pkg.Type.GetFormat(pkgInfo),
	})
}

func (inst *Installer) GetFilePath(pkg *config.Package, file *registry.File) (string, error) {
	u, err := inst.getURL(pkg)
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}

	src, err := pkg.RenderSrc(file, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("get a file src: %w", err)
	}
	return filepath.Join(inst.getInstallDir(u), src), nil
}

func (inst *Installer) GetFormat(pkg *registry.PackageInfo) string {
	return pkg.Format
}

func (inst *Installer) GetName(pkg *registry.PackageInfo) string {
	return pkg.GetName()
}
