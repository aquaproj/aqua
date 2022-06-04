package githubarchive

import (
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/apperr"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/github/archive"
	"github.com/spf13/afero"
)

const PkgType = "github_archive"

func New(param *config.Param, fs afero.Fs, githubArchive archive.Downloader) *Installer {
	return &Installer{
		fs:            fs,
		rootDir:       param.RootDir,
		githubArchive: githubArchive,
	}
}

type Installer struct {
	fs            afero.Fs
	rootDir       string
	githubArchive archive.Downloader
}

func (inst *Installer) validate(pkgInfo *registry.PackageInfo) error {
	if !pkgInfo.HasRepo() {
		return apperr.ErrRepoRequired
	}
	return nil
}

func (inst *Installer) getFileSrc(file *registry.File) string {
	if file.Src != "" {
		return file.Src
	}
	return file.Name
}

func (inst *Installer) GetFormat(pkg *registry.PackageInfo) string {
	return "tar.gz"
}

func (inst *Installer) GetFilePath(pkg *config.Package, file *registry.File) (string, error) {
	return filepath.Join(inst.getInstallDir(pkg), inst.getFileSrc(file)), nil
}
