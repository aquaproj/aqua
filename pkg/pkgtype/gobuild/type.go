package gobuild

import (
	"context"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/apperr"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/aquaproj/aqua/pkg/github/archive"
	"github.com/spf13/afero"
)

const PkgType = "go"

func New(param *config.Param, fs afero.Fs, builder GoBuilder, githubArchive archive.Downloader) *Installer {
	return &Installer{
		fs:            fs,
		rootDir:       param.RootDir,
		builder:       builder,
		githubArchive: githubArchive,
	}
}

type Installer struct {
	fs            afero.Fs
	rootDir       string
	builder       GoBuilder
	githubArchive archive.Downloader
}

type GoBuilder interface {
	GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error)
}

func NewGoBuilder(executor exec.Executor) GoBuilder {
	return executor
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
	return filepath.Join(inst.getBinDir(pkg, pkgInfo), inst.getFileSrc(file)), nil
}
