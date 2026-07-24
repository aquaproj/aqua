package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

// Abs roots p at dir if p starts with a slash, and returns p as it is
// otherwise. It lets a test case keep the readable absolute paths it was
// written with, such as /home/foo/aqua.yaml, while the files really live in a
// temporary directory. Relative paths, such as the path of a standard registry,
// are left alone.
func Abs(dir, p string) string {
	if !strings.HasPrefix(p, "/") {
		return p
	}
	return filepath.Join(dir, filepath.FromSlash(p))
}

// RootParam roots the paths of param at dir. See Abs.
func RootParam(dir string, param *config.Param) {
	param.CWD = Abs(dir, param.CWD)
	param.RootDir = Abs(dir, param.RootDir)
	param.File = Abs(dir, param.File)
	for i, p := range param.GlobalConfigFilePaths {
		param.GlobalConfigFilePaths[i] = Abs(dir, p)
	}
}

// RootEnv returns a copy of env whose PATH is rooted at dir. See Abs.
// The PATH of a test case is separated by colons, whatever the platform is.
func RootEnv(dir string, env map[string]string) map[string]string {
	m := make(map[string]string, len(env))
	for k, v := range env {
		if k != "PATH" {
			m[k] = v
			continue
		}
		paths := strings.Split(v, ":")
		for i, p := range paths {
			paths[i] = Abs(dir, p)
		}
		m[k] = strings.Join(paths, string(filepath.ListSeparator))
	}
	return m
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
