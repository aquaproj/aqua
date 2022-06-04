package http

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/apperr"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
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

func (inst *Installer) validate(pkgInfo *config.PackageInfo) error {
	if pkgInfo.Name == "" {
		return apperr.ErrPkgNameIsRequired
	}
	if pkgInfo.URL == nil {
		return errURLRequired
	}
	return nil
}

func (inst *Installer) GetFilePath(pkg *config.Package, pkgInfo *config.PackageInfo, file *config.File) (string, error) {
	uS, err := pkgInfo.RenderURL(pkg, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("render URL: %w", err)
	}

	u, err := url.Parse(uS)
	if err != nil {
		return "", fmt.Errorf("parse the URL: %w", err)
	}
	src, err := file.GetSrc(pkg, pkgInfo, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("get a file src: %w", err)
	}
	return filepath.Join(inst.getInstallDir(u), src), nil
}
