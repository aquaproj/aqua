package installpackage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *InstallerImpl) checkAndCopyFile(ctx context.Context, pkg *config.Package, file *registry.File, logE *logrus.Entry) error {
	exePath, err := inst.checkFileSrc(ctx, pkg, file, logE)
	if err != nil {
		return fmt.Errorf("check file_src is correct: %w", err)
	}
	if inst.copyDir == "" {
		return nil
	}
	logE.Info("copying an executable file")
	if err := inst.Copy(filepath.Join(inst.copyDir, file.Name), exePath); err != nil {
		return err
	}

	return nil
}

func (inst *InstallerImpl) checkFileSrcGo(ctx context.Context, pkg *config.Package, file *registry.File, logE *logrus.Entry) (string, error) {
	pkgInfo := pkg.PackageInfo
	exePath := filepath.Join(inst.rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "bin", file.Name)
	if isWindows(inst.runtime.GOOS) {
		exePath += ".exe"
	}
	dir, err := pkg.RenderDir(file, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("render file dir: %w", err)
	}
	exeDir := filepath.Join(inst.rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "src", dir)
	if _, err := inst.fs.Stat(exePath); err == nil {
		return exePath, nil
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
	if err := inst.goBuildInstaller.Install(ctx, exePath, exeDir, src); err != nil {
		return "", fmt.Errorf("build Go tool: %w", err)
	}
	return exePath, nil
}

func (inst *InstallerImpl) checkFileSrc(ctx context.Context, pkg *config.Package, file *registry.File, logE *logrus.Entry) (string, error) {
	if pkg.PackageInfo.Type == "go_build" {
		return inst.checkFileSrcGo(ctx, pkg, file, logE)
	}

	pkgPath, err := pkg.GetPkgPath(inst.rootDir, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("get the package install path: %w", err)
	}

	fileSrc, err := pkg.RenameFile(logE, inst.fs, pkgPath, file, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("get file_src: %w", err)
	}

	exePath := filepath.Join(pkgPath, fileSrc)
	finfo, err := inst.fs.Stat(exePath)
	if err != nil {
		return "", fmt.Errorf("exe_path isn't found: %w", logerr.WithFields(&config.FileNotFoundError{
			Err: err,
		}, logE.WithField("exe_path", exePath).Data))
	}
	if finfo.IsDir() {
		return "", logerr.WithFields(errExePathIsDirectory, logE.WithField("exe_path", exePath).Data) //nolint:wrapcheck
	}

	logE.Debug("check the permission")
	if mode := finfo.Mode().Perm(); !util.IsOwnerExecutable(mode) {
		logE.Debug("add the permission to execute the command")
		if err := inst.fs.Chmod(exePath, util.AllowOwnerExec(mode)); err != nil {
			return "", logerr.WithFields(errChmod, logE.Data) //nolint:wrapcheck
		}
	}

	return exePath, nil
}
