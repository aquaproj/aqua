package installpackage

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (is *Installer) createLinks(logger *slog.Logger, pkgs []*config.Package) bool {
	failed := false

	var aquaProxyPathOnWindows string
	if is.runtime.IsWindows() {
		pkg := proxyPkg()
		pkgPath, err := pkg.AbsPkgPath(is.rootDir, is.runtime)
		if err != nil {
			slogerr.WithError(logger, err).Error("get a path to aqua-proxy")
			failed = true
		}
		aquaProxyPathOnWindows = filepath.Join(pkgPath, "aqua-proxy.exe")
	}

	for _, pkg := range pkgs {
		logger := logger.With(
			"package_name", pkg.Package.Name,
			"package_version", pkg.Package.Version,
		)
		if is.createPackageLinks(logger, pkg, aquaProxyPathOnWindows) {
			failed = true
		}
	}
	return failed
}

func (is *Installer) createPackageLinks(logger *slog.Logger, pkg *config.Package, aquaProxyPathOnWindows string) bool {
	failed := false
	pkgInfo := pkg.PackageInfo
	for _, file := range pkgInfo.GetFiles() {
		logger := logger.With("command", file.Name)
		if is.createFileLinks(logger, pkg, file, aquaProxyPathOnWindows) {
			failed = true
		}
	}
	return failed
}

func (is *Installer) createFileLinks(logger *slog.Logger, pkg *config.Package, file *registry.File, aquaProxyPathOnWindows string) bool {
	failed := false
	cmds := map[string]struct{}{
		file.Name: {},
	}
	for _, alias := range pkg.Package.CommandAliases {
		if file.Name != alias.Command {
			continue
		}
		if alias.NoLink {
			continue
		}
		cmds[alias.Alias] = struct{}{}
	}
	for cmd := range cmds {
		if err := is.createCmdLink(logger, file, cmd, aquaProxyPathOnWindows); err != nil {
			slogerr.WithError(logger, err).Error("create a link to aqua-proxy")
			failed = true
		}
	}
	return failed
}

func (is *Installer) createCmdLink(logger *slog.Logger, file *registry.File, cmd string, aquaProxyPathOnWindows string) error {
	if cmd != file.Name {
		logger = logger.With("command_alias", cmd)
	}
	if is.realRuntime.IsWindows() {
		if err := is.createHardLinkToProxy(logger, cmd, aquaProxyPathOnWindows); err != nil {
			return fmt.Errorf("create a hard link to aqua-proxy: %w", err)
		}
		return nil
	}
	if err := is.createLink(logger, filepath.Join(is.rootDir, "bin", cmd), filepath.Join("..", proxyName)); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func (is *Installer) createHardLinkToProxy(logger *slog.Logger, cmd, aquaProxyPathOnWindows string) error {
	hardLink := filepath.Join(is.rootDir, "bin", cmd+exeExt)
	if f, err := afero.Exists(is.fs, hardLink); err != nil {
		return fmt.Errorf("check if a hard link to aqua-proxy exists: %w", err)
	} else if f {
		return nil
	}
	logger.Info("creating a hard link to aqua-proxy")
	if err := is.linker.Hardlink(aquaProxyPathOnWindows, hardLink); err != nil {
		return fmt.Errorf("create a hard link to aqua-proxy: %w", err)
	}
	return nil
}

func (is *Installer) recreateHardLinks() error {
	binDir := filepath.Join(is.rootDir, "bin")
	infos, err := afero.ReadDir(is.fs, binDir)
	if err != nil {
		return fmt.Errorf("read a bin dir: %w", err)
	}

	pkg := proxyPkg()
	pkgPath, err := pkg.AbsPkgPath(is.rootDir, is.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}
	a := filepath.Join(pkgPath, "aqua-proxy.exe")

	for _, info := range infos {
		if info.Name() == "aqua.exe" {
			continue
		}
		p := filepath.Join(binDir, info.Name())
		if err := is.fs.Remove(p); err != nil {
			return fmt.Errorf("remove a file to replace it with a hard link: %w", err)
		}
		if strings.HasSuffix(info.Name(), exeExt) {
			if err := is.linker.Hardlink(a, p); err != nil {
				return fmt.Errorf("create a hard link: %w", err)
			}
		}
	}
	return nil
}

func (is *Installer) createLink(logger *slog.Logger, linkPath, linkDest string) error {
	if fileInfo, err := is.linker.Lstat(linkPath); err == nil {
		switch mode := fileInfo.Mode(); {
		case mode.IsDir():
			// if file is a directory, raise error
			return fmt.Errorf("%s has already existed and is a directory", linkPath)
		case mode&os.ModeNamedPipe != 0:
			// if file is a pipe, raise error
			return fmt.Errorf("%s has already existed and is a named pipe", linkPath)
		case mode.IsRegular():
			// if file is a regular file, remove it and create a symlink.
			if err := is.fs.Remove(linkPath); err != nil {
				return fmt.Errorf("remove a file to create a symbolic link (%s): %w", linkPath, err)
			}
			if err := is.linker.Symlink(linkDest, linkPath); err != nil {
				return fmt.Errorf("create a symbolic link: %w", err)
			}
			return nil
		case mode&os.ModeSymlink != 0:
			return is.recreateLink(logger, linkPath, linkDest)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", linkPath, mode.String())
		}
	}
	logger.Info("create a symbolic link")
	if err := is.linker.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func (is *Installer) recreateLink(logger *slog.Logger, linkPath, linkDest string) error {
	lnDest, err := is.linker.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("read a symbolic link (%s): %w", linkPath, err)
	}
	if linkDest == lnDest {
		return nil
	}
	// recreate link
	logger.Debug("recreate a symbolic link",
		// TODO add version
		"link_file", linkPath,
		"old", lnDest,
		"new", linkDest)
	if err := is.fs.Remove(linkPath); err != nil {
		return fmt.Errorf("remove a symbolic link (%s): %w", linkPath, err)
	}
	if err := is.linker.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}
