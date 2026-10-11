package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// A release publishing a musl build beside the Linux one gets a variant choosing it, and a
// release whose Linux build is already the musl one gets none.
func TestParseAssetInfos_musl(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		assets  []string
		version string
		want    map[string]string // "<arch>/<libc>" -> asset
	}{
		{
			name: "UsefulSoftwareCo/executor",
			assets: []string{
				"executor-darwin-arm64.zip", "executor-darwin-x64.zip",
				"executor-linux-arm64-musl.tar.gz", "executor-linux-arm64.tar.gz",
				"executor-linux-x64-musl.tar.gz", "executor-linux-x64.tar.gz",
				"executor-windows-arm64.zip", "executor-windows-x64.zip",
			},
			version: "v1.6.10",
			want: map[string]string{
				"amd64/glibc": "executor-linux-x64.tar.gz",
				"amd64/musl":  "executor-linux-x64-musl.tar.gz",
				"arm64/glibc": "executor-linux-arm64.tar.gz",
				"arm64/musl":  "executor-linux-arm64-musl.tar.gz",
			},
		},
		{
			// Only amd64 has a musl build, and it spells the architecture differently.
			name: "one architecture",
			assets: []string{
				"foo-linux-amd64.tar.gz", "foo-linux-arm64.tar.gz",
				"foo-linux-x86_64-musl.tar.gz",
				"foo-darwin-amd64.tar.gz", "foo-darwin-arm64.tar.gz",
			},
			version: "v1.0.0",
			want: map[string]string{
				"amd64/glibc": "foo-linux-amd64.tar.gz",
				"amd64/musl":  "foo-linux-x86_64-musl.tar.gz",
				"arm64/glibc": "foo-linux-arm64.tar.gz",
				"arm64/musl":  "foo-linux-arm64.tar.gz",
			},
		},
		{
			// The musl triple is chosen for Linux already, and runs on glibc too.
			name: "Rust target triples",
			assets: []string{
				"rg-15.2.0-x86_64-unknown-linux-musl.tar.gz", "rg-15.2.0-x86_64-unknown-linux-gnu.tar.gz",
				"rg-15.2.0-aarch64-unknown-linux-gnu.tar.gz",
				"rg-15.2.0-x86_64-apple-darwin.tar.gz", "rg-15.2.0-aarch64-apple-darwin.tar.gz",
			},
			version: "15.2.0",
			want: map[string]string{
				"amd64/glibc": "rg-15.2.0-x86_64-unknown-linux-musl.tar.gz",
				"amd64/musl":  "rg-15.2.0-x86_64-unknown-linux-musl.tar.gz",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			infos := make([]*asset.AssetInfo, 0, len(d.assets))
			for _, name := range d.assets {
				infos = append(infos, asset.ParseAssetName(name, d.version))
			}
			pkgInfo := &registry.PackageInfo{Type: "github_release", RepoOwner: "o", RepoName: "r"}
			asset.ParseAssetInfos(pkgInfo, infos)
			for _, goarch := range []string{"amd64", "arm64"} {
				for _, libc := range []string{"glibc", "musl"} {
					want, ok := d.want[goarch+"/"+libc]
					if !ok {
						continue
					}
					rt := &runtime.Runtime{GOOS: "linux", GOARCH: goarch, LibC: libc}
					p := pkgInfo.Copy()
					p.OverrideByRuntime(rt)
					pkg := &config.Package{PackageInfo: p, Package: &aqua.Package{Name: "o/r", Version: d.version}}
					got, err := pkg.RenderAsset(rt)
					if err != nil {
						t.Fatal(err)
					}
					if got != want {
						t.Errorf("linux/%s %s: got %s, want %s", goarch, libc, got, want)
					}
				}
			}
		})
	}
}
