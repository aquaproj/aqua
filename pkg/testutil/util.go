package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
)

const dirPermission os.FileMode = 0o775

func NewFs(files map[string]string, dirs ...string) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	for name, body := range files {
		if err := afero.WriteFile(fs, name, []byte(body), osfile.FilePermission); err != nil {
			return nil, err //nolint:wrapcheck
		}
	}
	for _, dir := range dirs {
		if err := fs.MkdirAll(dir, dirPermission); err != nil {
			return nil, err //nolint:wrapcheck
		}
	}
	return fs, nil
}

// WriteFiles creates the given files and directories in dir.
// The keys of files and the elements of dirs are slash separated paths relative
// to dir. Parent directories are created as needed.
func WriteFiles(t *testing.T, dir string, files map[string]string, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := osfile.MkdirAll(filepath.Join(dir, filepath.FromSlash(d))); err != nil {
			t.Fatal(err)
		}
	}
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := osfile.MkdirAll(filepath.Dir(p)); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), osfile.FilePermission); err != nil {
			t.Fatal(err)
		}
	}
}
