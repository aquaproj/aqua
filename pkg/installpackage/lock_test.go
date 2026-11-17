package installpackage_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/google/go-cmp/cmp"
)

func lockTestConfig(pkgs ...*aqua.Package) *aqua.Config {
	return &aqua.Config{
		Registries: aqua.Registries{"standard": {Name: "standard", Type: "standard", Ref: "v4.0.0"}},
		Packages:   pkgs,
	}
}

func lockTestPackage(name, version string) *aqua.Package {
	return &aqua.Package{Name: name, Version: version, Registry: "standard"}
}

func TestSplitByLockFile(t *testing.T) {
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	cfg := lockTestConfig(
		lockTestPackage("cli/cli", "v2.1.0"),
		lockTestPackage("golang.org/x/tools/gopls", "v0.14.0"),
	)
	lf := &lockfile.LockFile{
		Packages: []*lockfile.Entry{
			{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Checksum: "aaa", ChecksumAlgorithm: "sha256", Commands: []string{"gh"}},
			{Name: "golang.org/x/tools/gopls", Version: "v0.14.0", OS: "linux", Arch: "amd64", Commands: []string{"gopls"}},
		},
	}

	locked, rest, err := installpackage.SplitByLockFile(slog.New(slog.DiscardHandler), lf, cfg, rt)
	if err != nil {
		t.Fatal(err)
	}

	names := make([]string, len(locked))
	for i, target := range locked {
		names[i] = target.Pkg.Package.Name
	}
	if diff := cmp.Diff([]string{"cli/cli", "golang.org/x/tools/gopls"}, names); diff != "" {
		t.Errorf("the locked packages are wrong (-want +got):\n%s", diff)
	}
	// A lock file leaves nothing for a registry to resolve.
	if len(rest) != 0 {
		t.Errorf("got %d packages for the registries, want none", len(rest))
	}

	checkUnresolved(t, locked[0])
	// The registry the package came from stays on the package, because a policy is
	// written against it.
	if locked[0].Pkg.Registry == nil || locked[0].Pkg.Registry.Name != "standard" {
		t.Errorf("the registry is %+v, want the standard one", locked[0].Pkg.Registry)
	}
}

// checkUnresolved checks that nothing is read from a registry yet: a target knows its
// entry and its commands, which is what creating links needs, and nothing it could be
// downloaded from.
func checkUnresolved(t *testing.T, target *installpackage.Target) {
	t.Helper()
	if target.Entry == nil || target.Entry.Checksum != "aaa" {
		t.Errorf("the entry is %+v, want the one with checksum aaa", target.Entry)
	}
	info := target.Pkg.PackageInfo
	if len(info.Files) != 1 || info.Files[0].Name != "gh" || info.Files[0].Src != "" {
		t.Errorf("the files are %+v, want the command gh alone", info.Files)
	}
	if info.Type != "" || info.RepoOwner != "" || info.Asset != "" || info.URL != "" {
		t.Errorf("the definition names a download: %+v", info)
	}
}

// A repository with no lock file hasn't adopted one, so every package still resolves
// through a registry exactly as before.
func TestSplitByLockFile_noLockFile(t *testing.T) {
	t.Parallel()
	cfg := lockTestConfig(lockTestPackage("cli/cli", "v2.1.0"))
	locked, rest, err := installpackage.SplitByLockFile(slog.New(slog.DiscardHandler), nil, cfg, &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"})
	if err != nil {
		t.Fatal(err)
	}
	if len(locked) != 0 {
		t.Errorf("got %d locked packages, want none", len(locked))
	}
	if diff := cmp.Diff(cfg.Packages, rest); diff != "" {
		t.Errorf("the remaining packages are wrong (-want +got):\n%s", diff)
	}
}

// Once a lock file exists it is the only source. Falling back to a registry would
// install something the lock file never described while the file still claims to say
// what is installed.
func TestSplitByLockFile_missingEntry(t *testing.T) {
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	tests := []struct {
		name string
		pkg  *aqua.Package
		lf   *lockfile.LockFile
	}{
		{
			name: "no entry at all",
			pkg:  lockTestPackage("foo/foo", "v1.0.0"),
			lf:   &lockfile.LockFile{},
		},
		{
			// An entry written before the lock file recorded commands. Guessing
			// them from the package name would create the wrong links.
			name: "an entry without commands",
			pkg:  lockTestPackage("cli/cli", "v2.1.0"),
			lf: &lockfile.LockFile{Packages: []*lockfile.Entry{
				{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Checksum: "aaa"},
			}},
		},
		{
			// A command name becomes a link in $(aqua root-dir)/bin, so a path would
			// replace a file outside it.
			name: "a command that isn't a plain name",
			pkg:  lockTestPackage("cli/cli", "v2.1.0"),
			lf: &lockfile.LockFile{Packages: []*lockfile.Entry{
				{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Checksum: "aaa", Commands: []string{"../../../.bashrc"}},
			}},
		},
		{
			name: "a package with no version",
			pkg:  lockTestPackage("qux/qux", ""),
			lf:   &lockfile.LockFile{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			locked, rest, err := installpackage.SplitByLockFile(slog.New(slog.DiscardHandler), tt.lf, lockTestConfig(tt.pkg), rt)
			if err == nil {
				t.Fatal("an error must be returned")
			}
			if len(locked) != 0 || len(rest) != 0 {
				t.Errorf("got %d locked and %d remaining packages, want none", len(locked), len(rest))
			}
		})
	}
}

// A package the lock file describes but not for this machine has no build here. That
// is what supported_envs says, and aqua has always passed over such a package rather
// than failing on it, so it is neither installed nor sent to a registry.
func TestSplitByLockFile_unsupportedEnvironment(t *testing.T) {
	t.Parallel()
	cfg := lockTestConfig(lockTestPackage("baz/baz", "v1.0.0"))
	lf := &lockfile.LockFile{Packages: []*lockfile.Entry{
		{Name: "baz/baz", Version: "v1.0.0", OS: "darwin", Arch: "arm64", Checksum: "ccc"},
	}}

	locked, rest, err := installpackage.SplitByLockFile(slog.New(slog.DiscardHandler), lf, cfg, &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"})
	if err != nil {
		t.Fatal(err)
	}
	if len(locked) != 0 {
		t.Errorf("got %d locked packages, want none", len(locked))
	}
	// Not sent to a registry either: the lock file answered, and the answer is that
	// there is nothing to install here.
	if len(rest) != 0 {
		t.Errorf("got %d packages for the registries, want none", len(rest))
	}
}

// A locked package is installed from the registry's definition with the lock file's
// checksum. A type that downloads something must have one: installing it without would
// be a hole in the guarantee the lock file exists to make.
func TestNewTarget(t *testing.T) {
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	rgst := &aqua.Registry{Name: "standard", Type: "standard"}
	tests := []struct {
		name     string
		def      *lockfile.Package
		checksum string
		isErr    bool
	}{
		{
			name: "a checksum",
			def: &lockfile.Package{
				Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64",
				Type: "github_release", RepoOwner: "cli", RepoName: "cli", Asset: "gh.tar.gz", Format: "tar.gz",
				Checksum: "aaa", ChecksumAlgorithm: "sha256",
			},
			checksum: "aaa",
		},
		{
			name: "no checksum where the type needs one",
			def: &lockfile.Package{
				Name: "bar/bar", Version: "v1.0.0", OS: "linux", Arch: "amd64",
				Type: "github_release", RepoOwner: "bar", RepoName: "bar", Asset: "bar.tar.gz",
			},
			isErr: true,
		},
		{
			// go_install downloads nothing, so there is no checksum to carry.
			name: "go_install",
			def: &lockfile.Package{
				Name: "golang.org/x/tools/gopls", Version: "v0.14.0", OS: "linux", Arch: "amd64",
				Type: "go_install", Path: "golang.org/x/tools/gopls",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkg := lockTestPackage(tt.def.Name, tt.def.Version)
			target, err := installpackage.NewTarget(pkg, tt.def, rgst, rt)
			if err != nil {
				if !tt.isErr {
					t.Fatal(err)
				}
				return
			}
			if tt.isErr {
				t.Fatal("an error must be returned")
			}
			checkTarget(t, target, rgst, tt.checksum)
		})
	}
}

func checkTarget(t *testing.T, target *installpackage.Target, rgst *aqua.Registry, chksum string) {
	t.Helper()
	if !target.Locked {
		t.Error("the target must say it came from the lock file")
	}
	got := ""
	if target.Checksum != nil {
		got = target.Checksum.Checksum
	}
	if got != chksum {
		t.Errorf("the checksum is %q, want %q", got, chksum)
	}
	// The registry the package came from stays on the package, because a policy is
	// written against it.
	if target.Pkg.Registry != rgst {
		t.Errorf("the registry is %+v, want the standard one", target.Pkg.Registry)
	}
}
