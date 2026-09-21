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
		Packages: []*lockfile.Package{
			{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Type: "github_release", RepoOwner: "cli", RepoName: "cli", Asset: "gh.tar.gz", Format: "tar.gz", Checksum: "aaa", ChecksumAlgorithm: "sha256"},
			{Name: "golang.org/x/tools/gopls", Version: "v0.14.0", OS: "linux", Arch: "amd64", Type: "go_install", Path: "golang.org/x/tools/gopls"},
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

	if locked[0].Checksum == nil || locked[0].Checksum.Checksum != "aaa" {
		t.Errorf("the checksum is %+v, want aaa", locked[0].Checksum)
	}
	// go_install downloads nothing, so there is no checksum to carry.
	if locked[1].Checksum != nil {
		t.Errorf("the checksum is %+v, want nothing", locked[1].Checksum)
	}
	// The registry the package came from stays on the package, because a policy is
	// written against it.
	if locked[0].Pkg.Registry == nil || locked[0].Pkg.Registry.Name != "standard" {
		t.Errorf("the registry is %+v, want the standard one", locked[0].Pkg.Registry)
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
			// Another environment's entry is not this one's.
			name: "no entry for this environment",
			pkg:  lockTestPackage("baz/baz", "v1.0.0"),
			lf: &lockfile.LockFile{Packages: []*lockfile.Package{
				{Name: "baz/baz", Version: "v1.0.0", OS: "darwin", Arch: "arm64", Type: "github_release", Asset: "baz.tar.gz", Checksum: "ccc"},
			}},
		},
		{
			// An entry whose type downloads an artifact but carries no checksum is
			// wrong rather than absent, and installing it would be a hole in the
			// guarantee the lock file exists to make.
			name: "an entry without the checksum its type needs",
			pkg:  lockTestPackage("bar/bar", "v1.0.0"),
			lf: &lockfile.LockFile{Packages: []*lockfile.Package{
				{Name: "bar/bar", Version: "v1.0.0", OS: "linux", Arch: "amd64", Type: "github_release", RepoOwner: "bar", RepoName: "bar", Asset: "bar.tar.gz"},
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
