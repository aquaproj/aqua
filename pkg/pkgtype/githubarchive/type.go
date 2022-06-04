package githubarchive

import (
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/apperr"
	"github.com/aquaproj/aqua/pkg/config"
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

func (inst *Installer) validate(pkgInfo *config.PackageInfo) error {
	if !pkgInfo.HasRepo() {
		return apperr.ErrRepoRequired
	}
	return nil
}

func (inst *Installer) getFileSrc(file *config.File) string {
	if file.Src != "" {
		return file.Src
	}
	return file.Name
}

func (inst *Installer) GetFilePath(pkg *config.Package, pkgInfo *config.PackageInfo, file *config.File) (string, error) {
	return filepath.Join(inst.getInstallDir(pkg, pkgInfo), inst.getFileSrc(file)), nil
}
