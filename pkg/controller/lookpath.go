package controller

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func lookPath(exeName string) string {
	for _, p := range strings.Split(os.Getenv("PATH"), ":") {
		bin := filepath.Join(p, exeName)
		finfo, err := readLink(bin)
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

func readLink(p string) (os.FileInfo, error) {
	finfo, err := os.Lstat(p)
	if err != nil {
		return nil, fmt.Errorf("get a file stat (%s): %w", p, err)
	}
	if finfo.Name() == proxyName {
		return finfo, nil
	}
	if finfo.Mode()&fs.ModeSymlink != 0 {
		s, err := os.Readlink(p)
		if err != nil {
			return nil, fmt.Errorf("read a symbolic link (%s): %w", p, err)
		}
		if filepath.IsAbs(s) {
			return readLink(s)
		}
		return readLink(filepath.Join(filepath.Dir(p), s))
	}
	return finfo, nil
}
