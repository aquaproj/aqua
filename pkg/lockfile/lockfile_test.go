package lockfile_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/google/go-cmp/cmp"
)

// TestSort checks the order a lock file is written in. The entries are a flat list,
// so the same packages have to come out the same way every time or a regenerated
// file produces a diff that says nothing.
func TestSort(t *testing.T) {
	t.Parallel()
	lf := &lockfile.LockFile{Packages: []*lockfile.Package{
		{Name: "cli/cli", Version: "v2.2.0", OS: "linux", Arch: "amd64"},
		{Name: "aquaproj/aqua", Version: "v2.0.0", OS: "linux", Arch: "arm64"},
		{Name: "cli/cli", Version: "v2.1.0", OS: "windows", Arch: "amd64"},
		{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Variants: map[string]string{"libc": "musl"}},
		{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Variants: map[string]string{"libc": "glibc"}},
		{Name: "aquaproj/aqua", Version: "v2.0.0", OS: "darwin", Arch: "arm64", Files: []*lockfile.File{
			{Name: "b"}, {Name: "a"},
		}},
	}}
	lf.Sort()

	got := make([]string, 0, len(lf.Packages))
	for _, pkg := range lf.Packages {
		got = append(got, strings.Join([]string{pkg.Name, pkg.Version, pkg.OS, pkg.Arch, pkg.Variants["libc"]}, " "))
	}
	want := []string{
		"aquaproj/aqua v2.0.0 darwin arm64 ",
		"aquaproj/aqua v2.0.0 linux arm64 ",
		// The two linux/amd64 entries differ only by variant, so the variant breaks
		// the tie.
		"cli/cli v2.1.0 linux amd64 glibc",
		"cli/cli v2.1.0 linux amd64 musl",
		"cli/cli v2.1.0 windows amd64 ",
		"cli/cli v2.2.0 linux amd64 ",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the order is wrong (-want +got):\n%s", diff)
	}
	if names := lf.Packages[0].Files; len(names) == 2 && names[0].Name != "a" {
		t.Errorf("files should be sorted, got %s then %s", names[0].Name, names[1].Name)
	}
}

func TestHasAndRemove(t *testing.T) {
	t.Parallel()
	lf := &lockfile.LockFile{Packages: []*lockfile.Package{
		{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64"},
		{Name: "cli/cli", Version: "v2.1.0", OS: "darwin", Arch: "arm64"},
		{Name: "cli/cli", Version: "v2.2.0", OS: "linux", Arch: "amd64"},
	}}
	if !lf.Has("cli/cli", "v2.1.0") {
		t.Error("the version is in the file")
	}
	if lf.Has("cli/cli", "v9.9.9") {
		t.Error("the version isn't in the file")
	}
	// Removing a version drops every environment of it, since a version is present
	// for all of them or none.
	lf.Remove("cli/cli", "v2.1.0")
	if lf.Has("cli/cli", "v2.1.0") {
		t.Error("the version should be gone")
	}
	if len(lf.Packages) != 1 {
		t.Errorf("the other version should remain, got %d entries", len(lf.Packages))
	}
}

// TestReadFile_missing checks that a repository which has never run
// 'aqua lock update' reads as an empty lock file rather than an error, since
// building one is what that command does.
func TestReadFile_missing(t *testing.T) {
	t.Parallel()
	lf, err := lockfile.ReadFile(filepath.Join(t.TempDir(), "aqua-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if lf.SchemaVersion != lockfile.SchemaVersion {
		t.Errorf("a new lock file should carry the schema version, got %q", lf.SchemaVersion)
	}
}

func TestWriteAndRead(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "aqua-lock.json")
	want := lockfile.New()
	want.Packages = []*lockfile.Package{{
		Name:     "cli/cli",
		Version:  "v2.1.0",
		Registry: &lockfile.Registry{Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry-g2"},
		OS:       "linux", Arch: "amd64",
		Type: "github_release", RepoOwner: "cli", RepoName: "cli",
		Asset: "gh_2.1.0_linux_amd64.tar.gz", Format: "tar.gz",
		Checksum: "abc", ChecksumAlgorithm: "sha256",
		Files: []*lockfile.File{{Name: "gh", Src: "gh_2.1.0_linux_amd64/bin/gh"}},
	}}
	if err := lockfile.Write(path, want); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(b), "\n") {
		t.Error("the file should end with a newline")
	}
	got, err := lockfile.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the lock file didn't round trip (-want +got):\n%s", diff)
	}
}
