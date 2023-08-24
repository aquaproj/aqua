//go:build windows
// +build windows

package which

import (
	"os"
	"path/filepath"
	"strings"
)

func (c *ControllerImpl) listExts() []string {
	x := c.osenv.Getenv(`PATHEXT`)
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

func (c *ControllerImpl) lookPath(envPath, exeName string) string {
	binDir := filepath.Join(c.rootDir, "bin")
	batDir := filepath.Join(c.rootDir, "bat")
	exts := c.listExts()
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
