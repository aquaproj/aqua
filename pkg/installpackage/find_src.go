package installpackage

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func containPath(p string) bool {
	switch p {
	case "README", "README.md", "LICENSE":
		return false
	}
	return true
}

func (inst *InstallerImpl) walk(pkgPath string) ([]string, error) {
	paths := []string{}
	if err := filepath.WalkDir(pkgPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr
		}
		if d.Type().IsDir() {
			return nil
		}
		f, err := filepath.Rel(pkgPath, path)
		if err != nil {
			return fmt.Errorf("get a relative path: %w", err)
		}
		name := d.Name()
		if containPath(name) {
			paths = append(paths, f)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walks the file tree of the unarchived package: %w", err)
	}
	return paths, nil
}
