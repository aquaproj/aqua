package which

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const proxyName = "aqua-proxy"

func (ctrl *controller) lookPath(envPath, exeName string) string {
	for _, p := range strings.Split(envPath, ":") {
		bin := filepath.Join(p, exeName)
		finfo, err := ctrl.readLink(bin)
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

func (ctrl *controller) readLink(p string) (os.FileInfo, error) {
	finfo, err := ctrl.linker.Lstat(p)
	if err != nil {
		return nil, fmt.Errorf("get a file stat (%s): %w", p, err)
	}
	if finfo.Name() == proxyName {
		return finfo, nil
	}
	if finfo.Mode()&fs.ModeSymlink != 0 {
		s, err := ctrl.linker.Readlink(p)
		if err != nil {
			return nil, fmt.Errorf("read a symbolic link (%s): %w", p, err)
		}
		if filepath.IsAbs(s) {
			return ctrl.readLink(s)
		}
		return ctrl.readLink(filepath.Join(filepath.Dir(p), s))
	}
	return finfo, nil
}
