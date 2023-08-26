package osfile

import "path/filepath"

func Abs(wd, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(wd, p)
}
