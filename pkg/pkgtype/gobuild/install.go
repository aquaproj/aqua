package gobuild

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *Installer) Install(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, logE *logrus.Entry) error {
	if err := inst.validate(pkgInfo); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	body, err := inst.githubArchive.Download(ctx, pkg, pkgInfo)
	if err != nil {
		return fmt.Errorf("download a GitHub Repository Archive: %w", err)
	}
	defer body.Close()

	dest := inst.getInstallDir(pkg, pkgInfo)

	if err := unarchive.Unarchive(&unarchive.File{
		Body: body,
		Type: "tar.gz",
	}, dest, logE, inst.fs); err != nil {
		return fmt.Errorf("unarchive a tarball: %w", logerr.WithFields(err, logrus.Fields{
			"dest": dest,
		}))
	}
	for _, file := range inst.GetFiles(pkgInfo) {
		if err := inst.build(ctx, pkg, pkgInfo, file, logE); err != nil {
			return err
		}
	}
	return nil
}

func (inst *Installer) CheckInstalled(pkg *config.Package, pkgInfo *config.PackageInfo) (bool, error) {
	binDir := inst.getBinDir(pkg, pkgInfo)
	for _, file := range inst.GetFiles(pkgInfo) {
		f, err := util.ExistFile(inst.fs, filepath.Join(binDir, inst.getFileSrc(file)))
		if err != nil {
			return false, fmt.Errorf("check if a file is installed: %w", err)
		}
		if !f {
			return false, nil
		}
	}
	return true, nil
}

func (inst *Installer) GetFiles(pkgInfo *config.PackageInfo) []*config.File {
	if files := pkgInfo.GetFiles(); len(files) != 0 {
		return files
	}
	// TODO
	return nil
}

func (inst *Installer) renderDir(pkg *config.Package, file *config.File) (string, error) {
	return template.Execute(file.Dir, map[string]interface{}{ //nolint:wrapcheck
		"Version":  pkg.Version,
		"FileName": file.Name,
	})
}

func (inst *Installer) getInstallDir(pkg *config.Package, pkgInfo *config.PackageInfo) string {
	return filepath.Join(inst.rootDir, "pkgs", PkgType, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, "src")
}

func (inst *Installer) getBinDir(pkg *config.Package, pkgInfo *config.PackageInfo) string {
	return filepath.Join(inst.rootDir, "pkgs", PkgType, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, "bin")
}

func (inst *Installer) build(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, file *config.File, logE *logrus.Entry) error {
	exePath := filepath.Join(inst.getBinDir(pkg, pkgInfo), file.Name)
	dir, err := inst.renderDir(pkg, file)
	if err != nil {
		return fmt.Errorf("render file dir: %w", err)
	}
	exeDir := filepath.Join(inst.getInstallDir(pkg, pkgInfo), dir)
	if _, err := inst.fs.Stat(exePath); err == nil {
		return nil
	}
	src := file.Src
	if src == "" {
		src = "."
	}
	logE.WithFields(logrus.Fields{
		"exe_path":     exePath,
		"go_src":       src,
		"go_build_dir": exeDir,
	}).Info("building Go tool")
	if _, err := inst.builder.GoBuild(ctx, exePath, src, exeDir); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}
