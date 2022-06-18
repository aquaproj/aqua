package installpackage

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (inst *installer) createLink(linkPath, linkDest string, logE *logrus.Entry) error {
	if fileInfo, err := inst.linker.Lstat(linkPath); err == nil {
		switch mode := fileInfo.Mode(); {
		case mode.IsDir():
			// if file is a directory, raise error
			return fmt.Errorf("%s has already existed and is a directory", linkPath)
		case mode&os.ModeNamedPipe != 0:
			// if file is a pipe, raise error
			return fmt.Errorf("%s has already existed and is a named pipe", linkPath)
		case mode.IsRegular():
			// TODO if file is a regular file, remove it and create a symlink.
			return fmt.Errorf("%s has already existed and is a regular file", linkPath)
		case mode&os.ModeSymlink != 0:
			return inst.recreateLink(linkPath, linkDest, logE)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", linkPath, mode.String())
		}
	}
	logE.WithFields(logrus.Fields{
		"link_file": linkPath,
		"new":       linkDest,
	}).Info("create a symbolic link")
	if err := inst.linker.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func (inst *installer) recreateLink(linkPath, linkDest string, logE *logrus.Entry) error {
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

const proxyTemplate = `@echo off
aqua exec -- <COMMAND> %*
`
const proxyPermission os.FileMode = 0o755

func (inst *installer) createLinkWindows(proxyPath, binName string, logE *logrus.Entry) error {
	if fileInfo, err := inst.linker.Lstat(proxyPath); err == nil {
		switch mode := fileInfo.Mode(); {
		case mode.IsDir():
			// if file is a directory, raise error
			return fmt.Errorf("%s has already existed and is a directory", proxyPath)
		case mode&os.ModeNamedPipe != 0:
			// if file is a pipe, raise error
			return fmt.Errorf("%s has already existed and is a named pipe", proxyPath)
		case mode.IsRegular():
			// TODO check content
			return nil
		case mode&os.ModeSymlink != 0:
			if err := inst.fs.Remove(proxyPath); err != nil {
				return fmt.Errorf("remove a symbolic link (%s): %w", proxyPath, err)
			}
			return inst.createProxyWindows(proxyPath, binName, logE)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", proxyPath, mode.String())
		}
	}

	return inst.createProxyWindows(proxyPath, binName, logE)
}

func (inst *installer) createProxyWindows(proxyPath, binName string, logE *logrus.Entry) error {
	logE.WithFields(logrus.Fields{
		"proxy_path": proxyPath,
	}).Info("create a proxy file")
	if err := afero.WriteFile(inst.fs, proxyPath, []byte(strings.Replace(proxyTemplate, "<COMMAND>", binName, 1)), proxyPermission); err != nil {
		return fmt.Errorf("create a proxy file (%s): %w", proxyPath, err)
	}
	return nil
}
