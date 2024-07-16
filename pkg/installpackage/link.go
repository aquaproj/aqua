package installpackage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (is *Installer) createLinks(logE *logrus.Entry, pkgs []*config.Package) bool {
	failed := false

	var aquaProxyPathOnWindows string
	if is.runtime.IsWindows() {
		pkg := proxyPkg()
		pkgPath, err := pkg.PkgPath(is.rootDir, is.runtime)
		if err != nil {
			logerr.WithError(logE, err).Error("get a path to aqua-proxy")
			failed = true
		}
		aquaProxyPathOnWindows = filepath.Join(pkgPath, "aqua-proxy.exe")
	}

	for _, pkg := range pkgs {
		pkgInfo := pkg.PackageInfo
		for _, file := range pkgInfo.GetFiles() {
			if isWindows(is.runtime.GOOS) {
				hardLink := filepath.Join(is.rootDir, "bin", file.Name+".exe")
				if f, err := afero.Exists(is.fs, hardLink); err != nil {
					logerr.WithError(logE, err).WithFields(logrus.Fields{
						"command": file.Name,
					}).Error("check if a hard link to aqua-proxy exists")
					failed = true
					continue
				} else if f {
					continue
				}
				logE.WithFields(logrus.Fields{
					"command": file.Name,
				}).Info("creating a hard link to aqua-proxy")
				if err := is.linker.Hardlink(aquaProxyPathOnWindows, hardLink); err != nil {
					logerr.WithError(logE, err).WithFields(logrus.Fields{
						"command": file.Name,
					}).Error("create a hard link to aqua-proxy")
					failed = true
				}
				continue
			}
			if err := is.createLink(filepath.Join(is.rootDir, "bin", file.Name), filepath.Join("..", proxyName), logE); err != nil {
				logerr.WithError(logE, err).Error("create the symbolic link")
				failed = true
				continue
			}
		}
	}
	return failed
}

func (is *Installer) recreateHardLinks() error {
	binDir := filepath.Join(is.rootDir, "bin")
	infos, err := afero.ReadDir(is.fs, binDir)
	if err != nil {
		return fmt.Errorf("read a bin dir: %w", err)
	}

	pkg := proxyPkg()
	pkgPath, err := pkg.PkgPath(is.rootDir, is.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}
	a := filepath.Join(pkgPath, "aqua-proxy.exe")

	for _, info := range infos {
		p := filepath.Join(binDir, info.Name())
		if err := is.fs.Remove(p); err != nil {
			return fmt.Errorf("remove a file to replace it with a hard link: %w", err)
		}
		if strings.HasSuffix(info.Name(), ".exe") {
			if err := is.linker.Hardlink(a, p); err != nil {
				return fmt.Errorf("create a hard link: %w", err)
			}
		}
	}
	return nil
}

func (is *Installer) createLink(linkPath, linkDest string, logE *logrus.Entry) error {
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
			return is.recreateLink(linkPath, linkDest, logE)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", linkPath, mode.String())
		}
	}
	logE.WithFields(logrus.Fields{
		"command": filepath.Base(linkPath),
	}).Info("create a symbolic link")
	if err := is.linker.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func (is *Installer) recreateLink(linkPath, linkDest string, logE *logrus.Entry) error {
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
