package installpackage

import (
	"fmt"
	"os"

	"github.com/aquaproj/aqua/pkg/log"
	"github.com/sirupsen/logrus"
)

func (inst *installer) createLink(linkPath, linkDest string) error {
	if fileInfo, err := os.Lstat(linkPath); err == nil {
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
			return recreateLink(linkPath, linkDest)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", linkPath, mode.String())
		}
	}
	inst.logE().WithFields(logrus.Fields{
		"link_file": linkPath,
		"new":       linkDest,
	}).Info("create a symbolic link")
	if err := os.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func recreateLink(linkPath, linkDest string) error {
	lnDest, err := os.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("read a symbolic link (%s): %w", linkPath, err)
	}
	if linkDest == lnDest {
		return nil
	}
	// recreate link
	log.New().WithFields(logrus.Fields{
		// TODO add version
		"link_file": linkPath,
		"old":       lnDest,
		"new":       linkDest,
	}).Debug("recreate a symbolic link")
	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("remove a symbolic link (%s): %w", linkPath, err)
	}
	if err := os.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}
