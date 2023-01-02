//go:build windows
// +build windows

package which

import (
	"os"
	"path/filepath"
	"strings"
)

func (ctrl *ControllerImpl) listExts() []string {
	x := ctrl.osenv.Getenv(`PATHEXT`)
	if x == "" {
		return []string{"", ".com", ".exe", ".bat", ".cmd"}
	}
	exts := []string{""}
	for _, e := range strings.Split(strings.ToLower(x), `;`) {
		if e == "" {
			continue
		}
		if e[0] != '.' {
			e = "." + e
		}
		exts = append(exts, e)
	}
	return exts
}

func (ctrl *ControllerImpl) lookPath(envPath, exeName string) string {
	binDir := filepath.Join(ctrl.rootDir, "bin")
	batDir := filepath.Join(ctrl.rootDir, "bat")
	exts := ctrl.listExts()
	for _, p := range filepath.SplitList(envPath) {
		if p == binDir || p == batDir {
			continue
		}
		bin := filepath.Join(p, exeName)
		for _, ext := range exts {
			bin := bin + ext
			finfo, err := os.Stat(bin)
			if err != nil {
				continue
			}
			if finfo.IsDir() {
				continue
			}
			return bin
		}
	}
	return ""
}
