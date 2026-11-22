package lockfile_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/google/go-cmp/cmp"
)

// TestSort checks the order a lock file is written in. The entries are a flat list,
// so the same packages have to come out the same way every time or a regenerated
// file produces a diff that says nothing.
func TestSort(t *testing.T) {
	t.Parallel()
	lf := &lockfile.LockFile{Packages: []*lockfile.Entry{
		{Name: "cli/cli", Version: "v2.2.0", OS: "linux", Arch: "amd64"},
		{Name: "aquaproj/aqua", Version: "v2.0.0", OS: "linux", Arch: "arm64"},
		{Name: "cli/cli", Version: "v2.1.0", OS: "windows", Arch: "amd64"},
		{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Variants: map[string]string{"libc": "musl"}},
		{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Variants: map[string]string{"libc": "glibc"}},
		{Name: "aquaproj/aqua", Version: "v2.0.0", OS: "darwin", Arch: "arm64", Commands: []string{"b", "a", "b"}},
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
	if diff := cmp.Diff([]string{"a", "b"}, lf.Packages[0].Commands); diff != "" {
		t.Errorf("commands should be sorted without repeats (-want +got):\n%s", diff)
	}
	// A null entry or registry can't be compared, and isn't written.
	withNull := &lockfile.LockFile{
		Packages:   []*lockfile.Entry{nil, {Name: "a"}},
		Registries: []*lockfile.Registry{nil, {Type: "github_content"}, nil},
	}
	withNull.Sort()
	if len(withNull.Packages) != 1 || len(withNull.Registries) != 1 {
		t.Errorf("got %d entries and %d registries", len(withNull.Packages), len(withNull.Registries))
	}
}

func TestHasAndRemove(t *testing.T) {
	t.Parallel()
	lf := &lockfile.LockFile{Packages: []*lockfile.Entry{
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
// A repository with no lock file reads as nothing, not as an empty lock file. The
// difference decides whether aqua may still resolve a package through a registry.
func TestReadFile_missing(t *testing.T) {
	t.Parallel()
	lf, err := lockfile.ReadFile(filepath.Join(t.TempDir(), "aqua-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if lf != nil {
		t.Errorf("got %+v, want nothing", lf)
	}
}

func TestWriteAndRead(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "aqua-lock.json")
	want := lockfile.New()
	want.Packages = []*lockfile.Entry{{
		Name:     "cli/cli",
		Version:  "v2.1.0",
		Registry: &lockfile.Registry{Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry-g2"},
		OS:       "linux", Arch: "amd64",
		Checksum: "abc", ChecksumAlgorithm: "sha256",
		Commands: []string{"gh"},
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

// What a lock file is for is the packages a configuration asks for, so an entry for one it
// doesn't answers a question nobody puts.
func TestLockFile_Retain(t *testing.T) {
	t.Parallel()
	lf := &lockfile.LockFile{Packages: []*lockfile.Entry{
		{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64"},
		{Name: "cli/cli", Version: "v2.1.0", OS: "darwin", Arch: "arm64"},
		// The version before it was bumped.
		{Name: "cli/cli", Version: "v2.0.0", OS: "linux", Arch: "amd64"},
		// A package that was removed from aqua.yaml.
		{Name: "gone/gone", Version: "v1.0.0", OS: "linux", Arch: "amd64"},
		nil,
	}}
	declared := map[string]struct{}{"cli/cli\tv2.1.0": {}}

	got := lf.Retain(func(name, version string) bool {
		_, ok := declared[name+"\t"+version]
		return ok
	})
	// The two it doesn't ask for, and the null entry, which is the one thing in the file
	// that can't be looked up at all.
	if got != 3 {
		t.Errorf("dropped %d entries, want 3", got)
	}
	// Every environment of what is still asked for stays: the file carries them all so
	// that one repository resolves on every machine that shares it.
	if len(lf.Packages) != 2 {
		t.Fatalf("kept %d entries", len(lf.Packages))
	}
	for _, pkg := range lf.Packages {
		if pkg.Version != "v2.1.0" {
			t.Errorf("kept %s %s", pkg.Name, pkg.Version)
		}
	}
}

// An entry records which commands a package has and nothing of where they come from, so
// that changing the lock file can't point a command at another download.
func TestPackage_Entry(t *testing.T) {
	t.Parallel()
	pkg := &lockfile.Package{
		Name:     "cli/cli",
		Version:  "v2.1.0",
		Registry: &lockfile.Registry{Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry-g2"},
		OS:       "linux", Arch: "amd64",
		Variants:   map[string]string{"libc": "musl"},
		LinkedLibc: "static",
		Type:       "github_release", RepoOwner: "cli", RepoName: "cli",
		Asset: "gh_2.1.0_linux_amd64.tar.gz", Format: "tar.gz",
		Checksum: "abc", ChecksumAlgorithm: "sha256",
		Files: []*lockfile.File{
			{Name: "gh", Src: "gh_2.1.0_linux_amd64/bin/gh"},
			{Name: "gh", Src: "elsewhere/gh"},
			{Name: "a", Src: "a"},
		},
		Cosign: &registry.Cosign{},
	}
	want := &lockfile.Entry{
		Name:     "cli/cli",
		Version:  "v2.1.0",
		Registry: &lockfile.Registry{Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry-g2"},
		OS:       "linux", Arch: "amd64",
		Variants:   map[string]string{"libc": "musl"},
		LinkedLibc: "static",
		Checksum:   "abc", ChecksumAlgorithm: "sha256",
		Commands: []string{"a", "gh"},
	}
	if diff := cmp.Diff(want, pkg.Entry()); diff != "" {
		t.Errorf("the entry is wrong (-want +got):\n%s", diff)
	}
	// A definition without files has the command named after the package, and the
	// entry records it rather than nothing.
	noFiles := *pkg
	noFiles.Files = nil
	if got := noFiles.Entry().Commands; len(got) != 1 || got[0] != "cli" {
		t.Errorf("the commands are %v, want the default one", got)
	}
	if !want.Describes(pkg) {
		t.Error("the entry describes the package it came from")
	}
	other := *pkg
	other.Variants = map[string]string{"libc": "glibc"}
	if want.Describes(&other) {
		t.Error("an entry for musl doesn't describe the glibc build")
	}
}

// A command name becomes the name of a link in $(aqua root-dir)/bin, so anything but a
// plain name would let a changed lock file replace a file elsewhere.
func TestEntry_CheckCommands(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		commands []string
		isErr    error
	}{
		"plain names":          {commands: []string{"gh", "kubectl-oidc_login", "terraform.exe"}},
		"none":                 {isErr: lockfile.ErrNoCommands},
		"a path":               {commands: []string{"../../../.bashrc"}, isErr: lockfile.ErrInvalidCommand},
		"a slash":              {commands: []string{"bin/gh"}, isErr: lockfile.ErrInvalidCommand},
		"a backslash":          {commands: []string{`..\evil`}, isErr: lockfile.ErrInvalidCommand},
		"a drive":              {commands: []string{"C:evil"}, isErr: lockfile.ErrInvalidCommand},
		"the parent directory": {commands: []string{".."}, isErr: lockfile.ErrInvalidCommand},
		"empty":                {commands: []string{""}, isErr: lockfile.ErrInvalidCommand},
		"one bad among good":   {commands: []string{"gh", "../x"}, isErr: lockfile.ErrInvalidCommand},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := (&lockfile.Entry{Commands: tt.commands}).CheckCommands()
			if tt.isErr == nil {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if !errors.Is(err, tt.isErr) {
				t.Errorf("got %v, want %v", err, tt.isErr)
			}
		})
	}
}

func TestLockFile_SetRegistry(t *testing.T) {
	t.Parallel()
	lf := &lockfile.LockFile{}
	key := &lockfile.Registry{Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml"}
	if lf.FindRegistry(key) != nil {
		t.Fatal("nothing is recorded yet")
	}
	first := *key
	first.Checksum = "aaa"
	lf.SetRegistry(&first)
	second := *key
	second.Checksum = "bbb"
	lf.SetRegistry(&second)
	// Another ref is another file.
	other := *key
	other.Ref = "v2.0.0"
	if lf.FindRegistry(&other) != nil {
		t.Error("another ref must not be found")
	}
	if got := lf.FindRegistry(key); got == nil || got.Checksum != "bbb" || len(lf.Registries) != 1 {
		t.Errorf("got %+v of %d", got, len(lf.Registries))
	}
}

// A null entry describes nothing, and is dropped when the file is read so that no lookup
// has to step around it.
func TestRead_null(t *testing.T) {
	t.Parallel()
	lf, err := lockfile.Read(strings.NewReader(`{"packages": [null, {"name": "a", "version": "v1", "os": "linux", "arch": "amd64"}], "registries": [null]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(lf.Packages) != 1 || len(lf.Registries) != 0 {
		t.Errorf("got %d entries and %d registries", len(lf.Packages), len(lf.Registries))
	}
}

// The commands are compared with the registry's as a sorted list, so the same commands in
// another order, as a hand edit or a merge leaves them, read the same.
func TestRead_commands(t *testing.T) {
	t.Parallel()
	lf, err := lockfile.Read(strings.NewReader(`{"packages": [{"name": "a", "version": "v1", "os": "linux", "arch": "amd64", "commands": ["b", "a", "b"]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"a", "b"}, lf.Packages[0].Commands); diff != "" {
		t.Errorf("the commands are wrong (-want +got):\n%s", diff)
	}
}

// The commit a registry was recorded at is written and read back with its checksum.
func TestWriteAndRead_registryCommit(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), lockfile.FileName)
	lf := lockfile.New()
	lf.Registries = []*lockfile.Registry{{
		Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml",
		Commit: strings.Repeat("a", 40), Checksum: "abc", ChecksumAlgorithm: "sha256",
	}}
	if err := lockfile.Write(p, lf); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"commit": "`+strings.Repeat("a", 40)+`"`) {
		t.Errorf("the commit isn't written: %s", b)
	}
	got, err := lockfile.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(lf.Registries, got.Registries); diff != "" {
		t.Error(diff)
	}
}
