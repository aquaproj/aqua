package installpackage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (inst *Installer) createLink(linkPath, linkDest string, logE *logrus.Entry) error {
	if fileInfo, err := inst.linker.Lstat(linkPath); err == nil {
		switch mode := fileInfo.Mode(); {
		case mode.IsDir():
			// if file is a directory, raise error
			return fmt.Errorf("%s has already existed and is a directory", linkPath)
		case mode&os.ModeNamedPipe != 0:
			// if file is a pipe, raise error
			return fmt.Errorf("%s has already existed and is a named pipe", linkPath)
		case mode.IsRegular():
			if err := inst.fs.Remove(linkPath); err != nil {
				return fmt.Errorf("remove a file to create a symbolic link (%s): %w", linkPath, err)
			}
			if err := inst.linker.Symlink(linkDest, linkPath); err != nil {
				return fmt.Errorf("create a symbolic link: %w", err)
			}
			return nil
		case mode&os.ModeSymlink != 0:
			return inst.recreateLink(linkPath, linkDest, logE)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", linkPath, mode.String())
		}
	}
	logE.WithFields(logrus.Fields{
		"command": filepath.Base(linkPath),
	}).Info("create a symbolic link")
	if err := inst.linker.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func (inst *Installer) recreateLink(linkPath, linkDest string, logE *logrus.Entry) error {
	lnDest, err := inst.linker.Readlink(linkPath)
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
	if err := inst.fs.Remove(linkPath); err != nil {
		return fmt.Errorf("remove a symbolic link (%s): %w", linkPath, err)
	}
	if err := inst.linker.Symlink(linkDest, linkPath); err != nil {
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

func (inst *Installer) createProxyWindows(binName string, logE *logrus.Entry) error {
	if err := inst.createBinWindows(filepath.Join(inst.rootDir, "bin", binName), scrTemplate, logE); err != nil {
		return err
	}
	if err := inst.createBinWindows(filepath.Join(inst.rootDir, "bat", binName+".bat"), strings.Replace(batTemplate, "<COMMAND>", binName, 1), logE); err != nil {
		return err
	}
	return nil
}

func (inst *Installer) createBinWindows(binPath, binTxt string, logE *logrus.Entry) error {
	if fileInfo, err := inst.linker.Lstat(binPath); err == nil {
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
			if err := inst.fs.Remove(binPath); err != nil {
				return fmt.Errorf("remove a symbolic link (%s): %w", binPath, err)
			}
			return inst.writeBinWindows(binPath, binTxt, logE)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", binPath, mode.String())
		}
	}

	return inst.writeBinWindows(binPath, binTxt, logE)
}

func (inst *Installer) writeBinWindows(proxyPath, binTxt string, logE *logrus.Entry) error {
	logE.WithFields(logrus.Fields{
		"proxy_path": proxyPath,
	}).Info("create a proxy file")
	if err := afero.WriteFile(inst.fs, proxyPath, []byte(binTxt), proxyPermission); err != nil {
		return fmt.Errorf("create a proxy file (%s): %w", proxyPath, err)
	}
	return nil
}
