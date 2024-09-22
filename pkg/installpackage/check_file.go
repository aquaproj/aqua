package installpackage

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (is *Installer) checkFilesWrap(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackage, pkgPath string) error {
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo

	failed := false
	notFound := false
	for _, file := range pkgInfo.GetFiles() {
		logE := logE.WithField("file_name", file.Name)
		var errFileNotFound *config.FileNotFoundError
		if err := is.checkAndCopyFile(ctx, logE, pkg, file); err != nil {
			if errors.As(err, &errFileNotFound) {
				notFound = true
			}
			failed = true
			logerr.WithError(logE, err).Error("check file_src is correct")
		}
	}
	if notFound { //nolint:nestif
		paths, err := is.walk(pkgPath)
		if err != nil {
			logerr.WithError(logE, err).Warn("traverse the content of unarchived package")
		} else {
			if len(paths) > 30 { //nolint:mnd
				logE.Errorf("executable files aren't found\nFiles in the unarchived package (Only 30 files are shown):\n%s\n ", strings.Join(paths[:30], "\n"))
			} else {
				logE.Errorf("executable files aren't found\nFiles in the unarchived package:\n%s\n ", strings.Join(paths, "\n"))
			}
		}
	}
	if failed {
		return errors.New("check file_src is correct")
	}

	return nil
}

func (is *Installer) checkAndCopyFile(ctx context.Context, logE *logrus.Entry, pkg *config.Package, file *registry.File) error {
	exePath, err := is.checkFileSrc(ctx, logE, pkg, file)
	if err != nil {
		return fmt.Errorf("check file_src is correct: %w", err)
	}
	if is.copyDir == "" {
		return nil
	}
	logE.Info("copying an executable file")
	if err := is.Copy(filepath.Join(is.copyDir, file.Name), exePath); err != nil {
		return err
	}

	return nil
}

func (is *Installer) checkFileSrcGo(ctx context.Context, logE *logrus.Entry, pkg *config.Package, file *registry.File) (string, error) {
	pkgInfo := pkg.PackageInfo
	exePath := filepath.Join(is.rootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "bin", file.Name)
	if isWindows(is.runtime.GOOS) {
		exePath += ".exe"
	}
	dir, err := pkg.RenderDir(file, is.runtime)
	if err != nil {
		return "", fmt.Errorf("render file dir: %w", err)
	}
	exeDir := filepath.Join(is.rootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "src", dir)
	if _, err := is.fs.Stat(exePath); err == nil {
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
	if err := is.goBuildInstaller.Install(ctx, exePath, exeDir, src); err != nil {
		return "", fmt.Errorf("build Go tool: %w", err)
	}
	return exePath, nil
}

func (is *Installer) checkFileSrc(ctx context.Context, logE *logrus.Entry, pkg *config.Package, file *registry.File) (string, error) {
	if pkg.PackageInfo.Type == "go_build" {
		return is.checkFileSrcGo(ctx, logE, pkg, file)
	}

	pkgPath, err := pkg.PkgPath(is.rootDir, is.runtime)
	if err != nil {
		return "", fmt.Errorf("get the package install path: %w", err)
	}

	fileSrc, err := pkg.RenameFile(logE, is.fs, pkgPath, file, is.runtime)
	if err != nil {
		return "", fmt.Errorf("get file_src: %w", err)
	}

	exePath := filepath.Join(pkgPath, fileSrc)
	finfo, err := is.fs.Stat(exePath)
	if err != nil {
		return "", fmt.Errorf("exe_path isn't found: %w", logerr.WithFields(&config.FileNotFoundError{
			Err: err,
		}, logE.WithField("exe_path", exePath).Data))
	}
	if finfo.IsDir() {
		return "", logerr.WithFields(errExePathIsDirectory, logE.WithField("exe_path", exePath).Data) //nolint:wrapcheck
	}

	logE.Debug("check the permission")
	if mode := finfo.Mode().Perm(); !osfile.IsOwnerExecutable(mode) {
		logE.Debug("add the permission to execute the command")
		if err := is.fs.Chmod(exePath, osfile.AllowOwnerExec(mode)); err != nil {
			return "", logerr.WithFields(errChmod, logE.Data) //nolint:wrapcheck
		}
	}

	return exePath, nil
}
