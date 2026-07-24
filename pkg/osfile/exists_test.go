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
	if f, err := osfile.Exists(dir); err != nil {
		t.Fatal(err)
	} else if !f {
		t.Fatal("a directory must exist")
	}

	p := filepath.Join(dir, "foo.txt")
	if f, err := osfile.Exists(p); err != nil {
		t.Fatal(err)
	} else if f {
		t.Fatal("a file must not exist")
	}
	if err := os.WriteFile(p, []byte("foo"), 0o600); err != nil {
		t.Fatal(err)
	}
	if f, err := osfile.Exists(p); err != nil {
		t.Fatal(err)
	} else if !f {
		t.Fatal("a file must exist")
	}
}

// A file whose parent directory can't be searched is not a missing file: the
// error must be surfaced rather than reported as absent.
func TestExists_unreadableParent(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}

	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(sub, "foo.txt")
	if err := os.WriteFile(p, []byte("foo"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(sub, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(sub, 0o755)
	})
	if _, err := osfile.Exists(p); err == nil {
		t.Fatal("an error must be returned when the file can't be stat'd")
	}
}
