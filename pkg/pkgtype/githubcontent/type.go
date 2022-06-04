package githubcontent

import (
	"errors"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/apperr"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github/content"
	"github.com/spf13/afero"
)

const PkgType = "github_content"

func New(param *config.Param, fs afero.Fs, githubContent content.Client) *Installer {
	return &Installer{
		fs:            fs,
		rootDir:       param.RootDir,
		githubContent: githubContent,
	}
}

type Installer struct {
	fs            afero.Fs
	rootDir       string
	githubContent content.Client
}

var errGitHubContentRequirePath = errors.New("github_content package requires path")

func (inst *Installer) validate(pkgInfo *config.PackageInfo) error {
	if !pkgInfo.HasRepo() {
		return apperr.ErrRepoRequired
	}
	if pkgInfo.Path == nil {
		return errGitHubContentRequirePath
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
