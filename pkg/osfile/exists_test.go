package osfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

func TestExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if !osfile.Exists(dir) {
		t.Fatal("a directory must exist")
	}

	p := filepath.Join(dir, "foo.txt")
	if osfile.Exists(p) {
		t.Fatal("a file must not exist")
	}
	if err := os.WriteFile(p, []byte("foo"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !osfile.Exists(p) {
		t.Fatal("a file must exist")
	}
}
