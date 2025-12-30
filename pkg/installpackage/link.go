package installpackage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (is *Installer) createLinks(logE *logrus.Entry, pkgs []*config.Package) bool {
	failed := false

	var aquaProxyPathOnWindows string
	if is.runtime.IsWindows() {
		pkg := proxyPkg()
		pkgPath, err := pkg.AbsPkgPath(is.rootDir, is.runtime)
		if err != nil {
			logerr.WithError(logE, err).Error("get a path to aqua-proxy")
			failed = true
		}
		aquaProxyPathOnWindows = filepath.Join(pkgPath, "aqua-proxy.exe")
	}

	for _, pkg := range pkgs {
		logE := logE.WithFields(logrus.Fields{
			"package_name":    pkg.Package.Name,
			"package_version": pkg.Package.Version,
		})
		if is.createPackageLinks(logE, pkg, aquaProxyPathOnWindows) {
			failed = true
		}
	}
	return failed
}

func (is *Installer) createPackageLinks(logE *logrus.Entry, pkg *config.Package, aquaProxyPathOnWindows string) bool {
	failed := false
	pkgInfo := pkg.PackageInfo
	for _, file := range pkgInfo.GetFiles() {
		logE := logE.WithFields(logrus.Fields{
			"command": file.Name,
		})
		if is.createFileLinks(logE, pkg, file, aquaProxyPathOnWindows) {
			failed = true
		}
	}
	return failed
}

func (is *Installer) createFileLinks(logE *logrus.Entry, pkg *config.Package, file *registry.File, aquaProxyPathOnWindows string) bool {
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
		if err := is.createCmdLink(logE, file, cmd, aquaProxyPathOnWindows); err != nil {
			logerr.WithError(logE, err).Error("create a link to aqua-proxy")
			failed = true
		}
	}
	return failed
}

func (is *Installer) createCmdLink(logE *logrus.Entry, file *registry.File, cmd string, aquaProxyPathOnWindows string) error {
	if cmd != file.Name {
		logE = logE.WithFields(logrus.Fields{
			"command_alias": cmd,
		})
	}
	if is.realRuntime.IsWindows() {
		if err := is.createHardLinkToProxy(logE, cmd, aquaProxyPathOnWindows); err != nil {
			return fmt.Errorf("create a hard link to aqua-proxy: %w", err)
		}
		return nil
	}
	if err := is.createLink(logE, filepath.Join(is.rootDir, "bin", cmd), filepath.Join("..", proxyName)); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func (is *Installer) createHardLinkToProxy(logE *logrus.Entry, cmd, aquaProxyPathOnWindows string) error {
	hardLink := filepath.Join(is.rootDir, "bin", cmd+exeExt)
	if f, err := afero.Exists(is.fs, hardLink); err != nil {
		return fmt.Errorf("check if a hard link to aqua-proxy exists: %w", err)
	} else if f {
		return nil
	}
	logE.Info("creating a hard link to aqua-proxy")
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

func (is *Installer) createLink(logE *logrus.Entry, linkPath, linkDest string) error {
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
			if is.realRuntime.IsWindows() {
				if err := is.linker.Hardlink(linkDest, linkPath); err != nil {
					return fmt.Errorf("create a hard link: %w", err)
				}
				return nil
			}
			if err := is.linker.Symlink(linkDest, linkPath); err != nil {
				return fmt.Errorf("create a symbolic link: %w", err)
			}
			return nil
		case mode&os.ModeSymlink != 0:
			if is.realRuntime.IsWindows() {
				if err := is.fs.Remove(linkPath); err != nil {
					return fmt.Errorf("remove a file to create a symbolic link (%s): %w", linkPath, err)
				}
				if err := is.linker.Hardlink(linkDest, linkPath); err != nil {
					return fmt.Errorf("create a hard link: %w", err)
				}
				return nil
			}
			return is.recreateLink(logE, linkPath, linkDest)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", linkPath, mode.String())
		}
	}
	if is.realRuntime.IsWindows() {
		logE.Info("create a hard link")
		if err := is.linker.Hardlink(linkDest, linkPath); err != nil {
			return fmt.Errorf("create a hard link: %w", err)
		}
		return nil
	}
	logE.Info("create a symbolic link")
	if err := is.linker.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func (is *Installer) recreateLink(logE *logrus.Entry, linkPath, linkDest string) error {
	lnDest, err := is.linker.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("read a symbolic link (%s): %w", linkPath, err)
	}
	if linkDest == lnDest {
		return nil
	}
	// recreate link
	logE.WithFields(logrus.Fields{
		// TODO add version
		"link_file": linkPath,
		"old":       lnDest,
		"new":       linkDest,
	}).Debug("recreate a symbolic link")
	if err := is.fs.Remove(linkPath); err != nil {
		return fmt.Errorf("remove a symbolic link (%s): %w", linkPath, err)
	}
	if err := is.linker.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}
