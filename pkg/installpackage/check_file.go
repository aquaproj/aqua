package installpackage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

const exeExt = ".exe"

func (is *Installer) checkFilesWrap(ctx context.Context, logger *slog.Logger, param *ParamInstallPackage, pkgPath string) error {
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo

	failed := false
	notFound := false
	for _, file := range pkgInfo.GetFiles() {
		logger := logger.With("file_name", file.Name)
		var errFileNotFound *config.FileNotFoundError
		if err := is.checkAndCopyFile(ctx, logger, pkg, file); err != nil {
			if errors.As(err, &errFileNotFound) {
				notFound = true
			}
			failed = true
			slogerr.WithError(logger, err).Error("check file_src is correct")
		}
	}
	if notFound { //nolint:nestif
		paths, err := is.walk(pkgPath)
		if err != nil {
			slogerr.WithError(logger, err).Warn("traverse the content of unarchived package")
		} else {
			if len(paths) > 30 { //nolint:mnd
				logger.Error(fmt.Sprintf("executable files aren't found\nFiles in the unarchived package (Only 30 files are shown):\n%s\n ", strings.Join(paths[:30], "\n")))
			} else {
				logger.Error(fmt.Sprintf("executable files aren't found\nFiles in the unarchived package:\n%s\n ", strings.Join(paths, "\n")))
			}
		}
	}
	if failed {
		return errors.New("check file_src is correct")
	}

	return nil
}

func (is *Installer) checkAndCopyFile(ctx context.Context, logger *slog.Logger, pkg *config.Package, file *registry.File) error {
	exePath, err := is.checkFileSrc(ctx, logger, pkg, file)
	if err != nil {
		return fmt.Errorf("check file_src is correct: %w", err)
	}
	if is.copyDir == "" {
		return nil
	}
	logger.Info("copying an executable file")
	exeNames := map[string]struct{}{
		file.Name: {},
	}
	for _, alias := range pkg.Package.CommandAliases {
		if alias.Command == file.Name {
			exeNames[alias.Alias] = struct{}{}
		}
	}

	for exeName := range exeNames {
		p := filepath.Join(is.copyDir, exeName)
		if is.runtime.IsWindows() && filepath.Ext(exeName) == "" {
			p += exeExt
		}
		if err := is.Copy(p, exePath); err != nil {
			return err
		}
	}
	return nil
}

func (is *Installer) checkFileSrcGo(ctx context.Context, logger *slog.Logger, pkg *config.Package, file *registry.File) (string, error) {
	pkgInfo := pkg.PackageInfo
	exePath := filepath.Join(is.rootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "bin", file.Name)
	if is.runtime.IsWindows() {
		exePath += exeExt
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
	logger.Info("building Go tool",
		"exe_path", exePath,
		"go_src", src,
		"go_build_dir", exeDir)
	if err := is.goBuildInstaller.Install(ctx, exePath, exeDir, src); err != nil {
		return "", fmt.Errorf("build Go tool: %w", err)
	}
	return exePath, nil
}

func (is *Installer) checkFileSrc(ctx context.Context, logger *slog.Logger, pkg *config.Package, file *registry.File) (string, error) {
	if pkg.PackageInfo.Type == "go_build" {
		return is.checkFileSrcGo(ctx, logger, pkg, file)
	}

	pkgPath, err := pkg.AbsPkgPath(is.rootDir, is.runtime)
	if err != nil {
		return "", fmt.Errorf("get the package install path: %w", err)
	}

	fileSrc, err := pkg.RenameFile(logger, is.fs, pkgPath, file, is.runtime)
	if err != nil {
		return "", fmt.Errorf("get file_src: %w", err)
	}

	exePath := filepath.Join(pkgPath, fileSrc)
	finfo, err := is.fs.Stat(exePath)
	if err != nil {
		return "", fmt.Errorf("exe_path isn't found: %w", slogerr.With(&config.FileNotFoundError{
			Err: err,
		}, "exe_path", exePath))
	}
	if finfo.IsDir() {
		return "", slogerr.With(errExePathIsDirectory, "exe_path", exePath) //nolint:wrapcheck
	}

	logger.Debug("check the permission")
	if mode := finfo.Mode().Perm(); !osfile.IsOwnerExecutable(mode) {
		logger.Debug("add the permission to execute the command")
		if err := is.fs.Chmod(exePath, osfile.AllowOwnerExec(mode)); err != nil {
			return "", errChmod
		}
	}

	if err := is.createFileLink(logger, file, exePath); err != nil {
		return "", err
	}

	return exePath, nil
}

func (is *Installer) createFileHardLink(logger *slog.Logger, file *registry.File, exePath string) error {
	link := filepath.Join(filepath.Dir(exePath), file.Link)
	if is.runtime.IsWindows() && filepath.Ext(link) == "" {
		link += exeExt
	}
	if f, err := afero.Exists(is.fs, link); err != nil {
		return fmt.Errorf("check if a hardlink exists: %w", err)
	} else if f {
		// do nothing
		return nil
	}
	logger.Info("creating a hard link")
	if err := is.linker.Hardlink(exePath, link); err != nil {
		return fmt.Errorf("create a hard link: %w", err)
	}
	return nil
}

func (is *Installer) createFileLink(logger *slog.Logger, file *registry.File, exePath string) error {
	if file.Link == "" {
		return nil
	}
	if file.Hard || is.runtime.IsWindows() {
		return is.createFileHardLink(logger, file, exePath)
	}
	// file.Link is the relative path from exePath to the link
	link := filepath.Join(filepath.Dir(exePath), file.Link)
	dest, err := filepath.Rel(filepath.Dir(link), exePath)
	if err != nil {
		return fmt.Errorf("get a dest of file.Link: %w", err)
	}

	if err := is.createLink(logger, link, dest); err != nil {
		return fmt.Errorf("create the symbolic link: %w", err)
	}
	return nil
}
