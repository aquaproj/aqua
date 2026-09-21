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

func TestSplitByLockFile(t *testing.T) {
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	cfg := &aqua.Config{
		Registries: aqua.Registries{"standard": {Name: "standard", Type: "standard", Ref: "v4.0.0"}},
		Packages: []*aqua.Package{
			{Name: "cli/cli", Version: "v2.1.0", Registry: "standard"},
			{Name: "foo/foo", Version: "v1.0.0", Registry: "standard"},
			{Name: "golang.org/x/tools/gopls", Version: "v0.14.0", Registry: "standard"},
			// A lock file entry without the checksum its type requires is wrong
			// rather than absent, so it must not be installed from.
			{Name: "bar/bar", Version: "v1.0.0", Registry: "standard"},
			// Another environment's entry is not this one's.
			{Name: "baz/baz", Version: "v1.0.0", Registry: "standard"},
		},
	}
	lf := &lockfile.LockFile{
		Packages: []*lockfile.Package{
			{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Type: "github_release", RepoOwner: "cli", RepoName: "cli", Asset: "gh.tar.gz", Format: "tar.gz", Checksum: "aaa", ChecksumAlgorithm: "sha256"},
			{Name: "golang.org/x/tools/gopls", Version: "v0.14.0", OS: "linux", Arch: "amd64", Type: "go_install", Path: "golang.org/x/tools/gopls"},
			{Name: "bar/bar", Version: "v1.0.0", OS: "linux", Arch: "amd64", Type: "github_release", RepoOwner: "bar", RepoName: "bar", Asset: "bar.tar.gz"},
			{Name: "baz/baz", Version: "v1.0.0", OS: "darwin", Arch: "arm64", Type: "github_release", RepoOwner: "baz", RepoName: "baz", Asset: "baz.tar.gz", Checksum: "ccc"},
		},
	}

	locked, rest := installpackage.SplitByLockFile(slog.New(slog.DiscardHandler), lf, cfg, rt)

	lockedNames := make([]string, len(locked))
	for i, t := range locked {
		lockedNames[i] = t.Pkg.Package.Name
	}
	if diff := cmp.Diff([]string{"cli/cli", "golang.org/x/tools/gopls"}, lockedNames); diff != "" {
		t.Errorf("the locked packages are wrong (-want +got):\n%s", diff)
	}

	restNames := make([]string, len(rest))
	for i, p := range rest {
		restNames[i] = p.Name
	}
	if diff := cmp.Diff([]string{"foo/foo", "bar/bar", "baz/baz"}, restNames); diff != "" {
		t.Errorf("the remaining packages are wrong (-want +got):\n%s", diff)
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
