package githubrelease

import (
	"errors"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/apperr"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/github/release"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/spf13/afero"
)

const PkgType = "github_release"

func New(param *config.Param, fs afero.Fs, rt *runtime.Runtime, githubRelease release.Client) *Installer {
	return &Installer{
		fs:            fs,
		runtime:       rt,
		rootDir:       param.RootDir,
		githubRelease: githubRelease,
	}
}

type Installer struct {
	fs            afero.Fs
	runtime       *runtime.Runtime
	rootDir       string
	githubRelease release.Client
}

var errAssetRequired = errors.New("github_release package requires asset")

func (inst *Installer) validate(pkgInfo *registry.PackageInfo) error {
	if !pkgInfo.HasRepo() {
		return apperr.ErrRepoRequired
	}
	if pkgInfo.Asset == nil {
		return errAssetRequired
	}
	return nil
}

func (inst *Installer) getFileSrc(file *registry.File) string {
	if file.Src != "" {
		return file.Src
	}
	return file.Name
}

func (inst *Installer) GetFilePath(pkg *config.Package, file *registry.File) (string, error) {
	assetName, err := inst.assetName(pkg)
	if err != nil {
		return "", err
	}
	return filepath.Join(inst.getInstallDir(pkg, assetName), inst.getFileSrc(file)), nil
}

func (inst *Installer) GetFormat(pkg *registry.PackageInfo) string {
	return pkg.Format
}

func (inst *Installer) GetName(pkg *registry.PackageInfo) string {
	return pkg.GetName()
}
