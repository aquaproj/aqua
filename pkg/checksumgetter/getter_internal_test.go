package checksumgetter

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
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
		resolve.RuntimeKey(rtLinux):  pkgLinux,
		resolve.RuntimeKey(rtDarwin): pkgDarwin,
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

func TestHasAssetSignature(t *testing.T) { //nolint:funlen
	t.Parallel()
	enabled := true
	disabled := false
	data := []struct {
		name    string
		pkgInfo *registry.PackageInfo
		exp     bool
	}{
		{
			name:    "nothing is signed",
			pkgInfo: &registry.PackageInfo{Type: "github_release"},
		},
		{
			name: "a checksum file is signed but the asset isn't",
			// The signature covers the file listing the digests, which is enough to
			// trust them without ever fetching the asset.
			pkgInfo: &registry.PackageInfo{
				Type: "github_release",
				Checksum: &registry.Checksum{
					Type:   "github_release",
					Asset:  "checksums.txt",
					Cosign: &registry.Cosign{},
				},
			},
		},
		{
			name: "cosign",
			pkgInfo: &registry.PackageInfo{
				Type:   "github_release",
				Cosign: &registry.Cosign{Opts: []string{"--key", "cosign.pub"}},
			},
			exp: true,
		},
		{
			// Cosign with nothing to verify with is not a signature.
			name: "an empty cosign",
			pkgInfo: &registry.PackageInfo{
				Type:   "github_release",
				Cosign: &registry.Cosign{},
			},
		},
		{
			name: "slsa provenance",
			pkgInfo: &registry.PackageInfo{
				Type:           "github_release",
				SLSAProvenance: &registry.SLSAProvenance{Type: "github_release", Asset: &[]string{"multiple.intoto.jsonl"}[0]},
			},
			exp: true,
		},
		{
			name: "github artifact attestations",
			pkgInfo: &registry.PackageInfo{
				Type:                       "github_release",
				GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{},
			},
			exp: true,
		},
		{
			name: "minisign",
			pkgInfo: &registry.PackageInfo{
				Type:     "github_release",
				Minisign: &registry.Minisign{PublicKey: "RWQ"},
			},
			exp: true,
		},
		{
			name: "a signature the registry turned off",
			pkgInfo: &registry.PackageInfo{
				Type:   "github_release",
				Cosign: &registry.Cosign{Opts: []string{"--key", "cosign.pub"}, Enabled: &disabled},
			},
		},
		{
			name: "a signature the registry turned on",
			pkgInfo: &registry.PackageInfo{
				Type:   "github_release",
				Cosign: &registry.Cosign{Enabled: &enabled},
			},
			exp: true,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if got := hasAssetSignature(d.pkgInfo); got != d.exp {
				t.Fatalf("hasAssetSignature = %v, wanted %v", got, d.exp)
			}
		})
	}
}

// An asset on a GitLab instance can be named by the path its release link was created
// with, while the checksum file names the file. The checksum has to be recorded under the
// asset, because that is what installing it then looks for.
func TestGetter_getChecksumsFromChecksumFile_nestedAsset(t *testing.T) {
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    "gitlab.com/ns/proj",
			Version: "v1.0.0",
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "gitlab_release",
			RepoOwner: "ns",
			RepoName:  "proj",
			Asset:     "binaries/tool-{{.OS}}-{{.Arch}}",
			Format:    "raw",
			Checksum: &registry.Checksum{
				Type:      "gitlab_release",
				Asset:     "checksums.txt",
				Algorithm: "sha256",
			},
		},
	}
	assetName, err := pkg.RenderAsset(rt)
	if err != nil {
		t.Fatalf("RenderAsset: %v", err)
	}
	if assetName != "binaries/tool-linux-amd64" {
		t.Fatalf("the asset is %q", assetName)
	}
	wantID, err := pkg.ChecksumID(rt)
	if err != nil {
		t.Fatalf("ChecksumID: %v", err)
	}

	g := &Getter{}
	const want = "89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101"
	cs, err := g.getChecksumsFromChecksumFile(pkg,
		map[string]struct{}{assetName: {}}, wantID,
		want+"  tool-linux-amd64\n"+
			"1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4  tool-darwin-arm64\n")
	if err != nil {
		t.Fatalf("getChecksumsFromChecksumFile: %v", err)
	}
	if len(cs) != 1 {
		t.Fatalf("wanted one checksum, got %d", len(cs))
	}
	if cs[0].ID != wantID {
		t.Fatalf("wanted the checksum to be recorded as %q, got %q", wantID, cs[0].ID)
	}
	if cs[0].Checksum != want {
		t.Fatalf("wanted %q, got %q", want, cs[0].Checksum)
	}
}
