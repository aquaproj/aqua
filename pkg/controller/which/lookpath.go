//go:build !windows
// +build !windows

package which

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const proxyName = "aqua-proxy"

func (c *ControllerImpl) lookPath(envPath, exeName string) string {
	for _, p := range filepath.SplitList(envPath) {
		bin := filepath.Join(p, exeName)
		finfo, err := c.readLink(bin)
		if err != nil {
			continue
		}
		if finfo.IsDir() {
			continue
		}
		if filepath.Base(finfo.Name()) == proxyName {
			continue
		}
		return bin
	}
	return ""
}

func (c *ControllerImpl) readLink(p string) (os.FileInfo, error) {
	finfo, err := c.linker.Lstat(p)
	if err != nil {
		return nil, fmt.Errorf("get a file stat (%s): %w", p, err)
	}
	if finfo.Name() == proxyName {
		return finfo, nil
	}
	if finfo.Mode()&fs.ModeSymlink != 0 {
		s, err := c.linker.Readlink(p)
		if err != nil {
			return nil, fmt.Errorf("read a symbolic link (%s): %w", p, err)
		}
		if filepath.IsAbs(s) {
			return c.readLink(s)
		}
		return c.readLink(filepath.Join(filepath.Dir(p), s))
	}
	return finfo, nil
}
