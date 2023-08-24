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

func (is *InstallerImpl) createLinks(logE *logrus.Entry, pkgs []*config.Package) bool {
	failed := false
	for _, pkg := range pkgs {
		pkgInfo := pkg.PackageInfo
		for _, file := range pkgInfo.GetFiles() {
			if isWindows(is.runtime.GOOS) {
				if err := is.createProxyWindows(file.Name, logE); err != nil {
					logerr.WithError(logE, err).Error("create the proxy file")
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

func (is *InstallerImpl) createLink(linkPath, linkDest string, logE *logrus.Entry) error {
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

func (is *InstallerImpl) recreateLink(linkPath, linkDest string, logE *logrus.Entry) error {
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

const (
	batTemplate = `@echo off
aqua exec -- <COMMAND> %*
`
	scrTemplate = `#!/usr/bin/env bash
exec aqua exec -- $0 $@
`
	proxyPermission os.FileMode = 0o755
)

func (is *InstallerImpl) createProxyWindows(binName string, logE *logrus.Entry) error {
	if err := is.createBinWindows(filepath.Join(is.rootDir, "bin", binName), scrTemplate, logE); err != nil {
		return err
	}
	if err := is.createBinWindows(filepath.Join(is.rootDir, "bat", binName+".bat"), strings.Replace(batTemplate, "<COMMAND>", binName, 1), logE); err != nil {
		return err
	}
	return nil
}

func (is *InstallerImpl) createBinWindows(binPath, binTxt string, logE *logrus.Entry) error {
	if fileInfo, err := is.linker.Lstat(binPath); err == nil {
		switch mode := fileInfo.Mode(); {
		case mode.IsDir():
			// if file is a directory, raise error
			return fmt.Errorf("%s has already existed and is a directory", binPath)
		case mode&os.ModeNamedPipe != 0:
			// if file is a pipe, raise error
			return fmt.Errorf("%s has already existed and is a named pipe", binPath)
		case mode.IsRegular():
			// TODO check content
			return nil
		case mode&os.ModeSymlink != 0:
			if err := is.fs.Remove(binPath); err != nil {
				return fmt.Errorf("remove a symbolic link (%s): %w", binPath, err)
			}
			return is.writeBinWindows(binPath, binTxt, logE)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", binPath, mode.String())
		}
	}

	return is.writeBinWindows(binPath, binTxt, logE)
}

func (is *InstallerImpl) writeBinWindows(proxyPath, binTxt string, logE *logrus.Entry) error {
	logE.WithFields(logrus.Fields{
		"proxy_path": proxyPath,
	}).Info("create a proxy file")
	if err := afero.WriteFile(is.fs, proxyPath, []byte(binTxt), proxyPermission); err != nil {
		return fmt.Errorf("create a proxy file (%s): %w", proxyPath, err)
	}
	return nil
}
