package goinstall

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/spf13/afero"
)

const PkgType = "go_install"

func New(param *config.Param, inst GoInstaller, fs afero.Fs) *Installer {
	return &Installer{
		fs:        fs,
		rootDir:   param.RootDir,
		installer: inst,
	}
}

type Installer struct {
	fs        afero.Fs
	rootDir   string
	installer GoInstaller
}

type GoInstaller interface {
	GoInstall(ctx context.Context, path, gobin string) (int, error)
}

func NewGoInstaller(executor exec.Executor) GoInstaller {
	return executor
}

var errGoInstallRequirePath = errors.New("go_install package requires path")

func (inst *Installer) validate(pkg *config.Package, pkgInfo *config.PackageInfo) error {
	if inst.getGoPath(pkgInfo) == "" {
		return errGoInstallRequirePath
	}
	if pkg.Version == "latest" {
		return errGoInstallForbidLatest
	}
	return nil
}

func (inst *Installer) getGoPath(pkgInfo *config.PackageInfo) string {
	if pkgInfo.Path != nil {
		return *pkgInfo.Path
	}
	if pkgInfo.HasRepo() {
		return "github.com/" + pkgInfo.RepoOwner + "/" + pkgInfo.RepoName
	}
	return ""
}

func (inst *Installer) getBinNames(pkgInfo *config.PackageInfo) []string {
	fileLen := len(pkgInfo.Files)
	if fileLen == 0 {
		return []string{filepath.Base(inst.getGoPath(pkgInfo))}
	}
	names := make([]string, fileLen)
	for i := 0; i < fileLen; i++ {
		names[i] = pkgInfo.Files[i].Name
	}
	return names
}

func (inst *Installer) getBinDir(pkg *config.Package, pkgInfo *config.PackageInfo) string {
	return filepath.Join(inst.rootDir, "pkgs", PkgType, inst.getGoPath(pkgInfo), pkg.Version, "bin")
}

func (inst *Installer) exist(binDir string, binNames []string) bool {
	for _, binName := range binNames {
		binPath := filepath.Join(binDir, binName)
		if _, err := inst.fs.Stat(binPath); err != nil {
			return false
		}
	}
	return true
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
