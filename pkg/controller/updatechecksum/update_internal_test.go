package updatechecksum

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func newGitHubReleasePkg(t *testing.T, owner, repoName, version, asset string, rt *runtime.Runtime) (*config.Package, string) {
	t.Helper()
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    owner + "/" + repoName,
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: owner,
			RepoName:  repoName,
			Asset:     asset,
		},
	}
	id, err := pkg.ChecksumID(rt)
	if err != nil {
		t.Fatalf("ChecksumID: %v", err)
	}
	return pkg, id
}

func TestAllChecksumsCached(t *testing.T) { //nolint:funlen
	t.Parallel()
	rtLinux := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	rtDarwin := &runtime.Runtime{GOOS: "darwin", GOARCH: "arm64"}

	pkgLinux, idLinux := newGitHubReleasePkg(t, "cli", "cli", "v2.17.0", "gh_linux_amd64.tar.gz", rtLinux)
	pkgDarwin, idDarwin := newGitHubReleasePkg(t, "cli", "cli", "v2.17.0", "gh_darwin_arm64.tar.gz", rtDarwin)

	pkgs := map[string]*config.Package{
		runtimeKey(rtLinux):  pkgLinux,
		runtimeKey(rtDarwin): pkgDarwin,
	}
	rts := []*runtime.Runtime{rtLinux, rtDarwin}

	const algo = "sha256"
	mkSum := func(id string) *checksum.Checksum {
		return &checksum.Checksum{ID: id, Checksum: "x", Algorithm: algo}
	}

	t.Run("empty rts is vacuously true", func(t *testing.T) {
		t.Parallel()
		if !allChecksumsCached(checksum.New(), pkgs, nil) {
			t.Fatal("want true for empty rts")
		}
	})

	t.Run("all cached", func(t *testing.T) {
		t.Parallel()
		c := checksum.New()
		c.Set(idLinux, mkSum(idLinux))
		c.Set(idDarwin, mkSum(idDarwin))
		if !allChecksumsCached(c, pkgs, rts) {
			t.Fatal("want true when every runtime checksum is present")
		}
	})

	t.Run("none cached", func(t *testing.T) {
		t.Parallel()
		if allChecksumsCached(checksum.New(), pkgs, rts) {
			t.Fatal("want false when no runtime checksum is present")
		}
	})

	t.Run("partial cached returns false", func(t *testing.T) {
		t.Parallel()
		c := checksum.New()
		c.Set(idLinux, mkSum(idLinux))
		// darwin missing
		if allChecksumsCached(c, pkgs, rts) {
			t.Fatal("want false when one runtime checksum is missing")
		}
	})

	t.Run("runtime missing from pkgs is skipped", func(t *testing.T) {
		t.Parallel()
		rtWindows := &runtime.Runtime{GOOS: "windows", GOARCH: "amd64"}
		c := checksum.New()
		c.Set(idLinux, mkSum(idLinux))
		c.Set(idDarwin, mkSum(idDarwin))
		// rtWindows has no entry in pkgs (e.g. cargo/go_install filtered by getPkgs);
		// it must not cause false.
		if !allChecksumsCached(c, pkgs, []*runtime.Runtime{rtLinux, rtDarwin, rtWindows}) {
			t.Fatal("runtime missing from pkgs should be skipped, not counted as missing")
		}
	})
}
