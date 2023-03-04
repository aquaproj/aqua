package installpackage

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
)

func containPath(p string) bool {
	switch p {
	case "README", "README.md", "LICENSE":
		return false
	}
	return true
}

func StringFileNotFoundErrors(pkg *config.Package, fileNotFoundErrors []*ErrorFileNotFound) string {
	msg := `Files aren't found.
Files:
%s

Maybe:
%s
`
	for _, fileNotFoundError := range fileNotFoundErrors {
		fmt.Sprintf(`name: %s\n

`)
	}
	return fmt.Sprintf(`files aren't found.`)
}

type ErrorFileNotFound struct {
	err  error
	msg  string
	file *registry.File
	src  string
}

func (err *ErrorFileNotFound) Error() string {
	return ""
}

func (inst *InstallerImpl) walk(pkgPath string, pkg *config.Package) ([]string, []*registry.File, error) {
	paths := []string{}
	candidates := make(map[string]*registry.File, len(pkg.PackageInfo.Files))
	for _, file := range pkg.PackageInfo.Files {
		candidates[file.Name] = &registry.File{
			Name: file.Name,
			Src:  file.Name,
		}
	}
	maybes := []*registry.File{}
	if err := filepath.WalkDir(pkgPath, func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsDir() {
			return nil
		}
		f, err := filepath.Rel(pkgPath, path)
		if err != nil {
			return err
		}
		name := d.Name()
		if containPath(name) {
			paths = append(paths, f)
			if file, ok := candidates[name]; ok {
				maybes = append(maybes, &registry.File{
					Name: file.Name,
					Src:  filepath.Join(filepath.Dir(f), file.Src),
				})
			}
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return paths, maybes, nil
}
