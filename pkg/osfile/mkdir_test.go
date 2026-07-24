package osfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

func TestMkdirAll(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "foo", "bar")
	if err := osfile.MkdirAll(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("a directory isn't created")
	}
	// An existing directory isn't an error.
	if err := osfile.MkdirAll(dir); err != nil {
		t.Fatal(err)
	}
}

// A temporary directory created by os.MkdirTemp would be 0700, which leaves a
// package renamed out of it unreadable to every other user.
// See https://github.com/aquaproj/aqua/issues/5049
func TestMkdirTemp_permission(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	got, err := osfile.MkdirTemp(parent)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(got) != parent {
		t.Fatalf("a directory is created in %s, want %s", filepath.Dir(got), parent)
	}

	// The expectation is a directory created the way aqua creates its own, so
	// that neither the umask nor the platform's handling of modes matters.
	want := filepath.Join(parent, "reference")
	if err := osfile.MkdirAll(want); err != nil {
		t.Fatal(err)
	}
	wantInfo, err := os.Stat(want)
	if err != nil {
		t.Fatal(err)
	}
	gotInfo, err := os.Stat(got)
	if err != nil {
		t.Fatal(err)
	}
	if gotInfo.Mode().Perm() != wantInfo.Mode().Perm() {
		t.Fatalf("the permission of the temporary directory is %o, want %o", gotInfo.Mode().Perm(), wantInfo.Mode().Perm())
	}
}

func TestMkdirTemp_unique(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	first, err := osfile.MkdirTemp(parent)
	if err != nil {
		t.Fatal(err)
	}
	second, err := osfile.MkdirTemp(parent)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("two temporary directories have the same path %s", first)
	}
}

func TestMkdirTemp_parentNotFound(t *testing.T) {
	t.Parallel()

	// The parent directory must exist. MkdirTemp creates only the leaf, so that
	// a typo in the parent path is an error rather than a stray directory tree.
	if _, err := osfile.MkdirTemp(filepath.Join(t.TempDir(), "not-found")); err == nil {
		t.Fatal("an error must occur")
	}
}
