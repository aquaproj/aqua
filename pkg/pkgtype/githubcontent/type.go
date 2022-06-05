package githubcontent

import (
	"errors"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/apperr"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
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

func (inst *Installer) validate(pkgInfo *registry.PackageInfo) error {
	if !pkgInfo.HasRepo() {
		return apperr.ErrRepoRequired
	}
	if pkgInfo.Path == nil {
		return errGitHubContentRequirePath
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
	return filepath.Join(inst.getInstallDir(pkg), inst.getFileSrc(file)), nil
}

func (inst *Installer) GetFormat(pkg *registry.PackageInfo) string {
	return pkg.Format
}

func (inst *Installer) GetName(pkg *registry.PackageInfo) string {
	return pkg.GetName()
}
