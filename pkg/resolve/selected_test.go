package resolve_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
	"github.com/google/go-cmp/cmp"
)

// The entry names the chosen asset as the release lists it, and what the definition still
// says -- the executables and the signature -- is rendered with the asset's own spellings
// and, for an emulated environment, with the amd64 build's.
func TestResolveSelected(t *testing.T) { //nolint:funlen
	t.Parallel()
	pkgInfo := &registry.PackageInfo{
		Name:      "foo/foo",
		Type:      "github_release",
		RepoOwner: "foo",
		RepoName:  "foo",
		// What the definition says about the asset no longer applies.
		Asset:  "foo-{{.OS}}-{{.Arch}}.{{.Format}}",
		Format: "zip",
		Files:  []*registry.File{{Name: "foo"}},
		SLSAProvenance: &registry.SLSAProvenance{
			Type:  "github_release",
			Asset: new("foo-{{.OS}}-{{.Arch}}.intoto.jsonl"),
		},
		Overrides: []*registry.Override{
			{
				GOOS:  "linux",
				Asset: "foo-linux-{{.Arch}}-old.tar.gz",
				Files: []*registry.File{{Name: "foo", Src: "bin/foo"}},
			},
		},
	}
	data := []struct {
		name string
		sel  []string
		want *lockfile.Package
	}{
		{
			name: "linux musl",
			sel:  []string{"foo-x86_64-unknown-linux-musl.tar.gz"},
			want: &lockfile.Package{
				Name: "foo/foo", Version: "v1.0.0", OS: "linux", Arch: "amd64",
				Variants: map[string]string{"libc": "musl"},
				Type:     "github_release", RepoOwner: "foo", RepoName: "foo",
				Asset:  "foo-x86_64-unknown-linux-musl.tar.gz",
				Format: "tar.gz",
				Files:  []*lockfile.File{{Name: "foo", Src: "bin/foo"}},
				SLSAProvenance: &registry.SLSAProvenance{
					Type:  "github_release",
					Asset: new("foo-unknown-linux-musl-x86_64.intoto.jsonl"),
				},
			},
		},
		{
			name: "Rosetta 2",
			sel:  []string{"foo-darwin-x64.zip"},
			want: &lockfile.Package{
				Name: "foo/foo", Version: "v1.0.0", OS: "darwin", Arch: "arm64",
				Type: "github_release", RepoOwner: "foo", RepoName: "foo",
				Asset:  "foo-darwin-x64.zip",
				Format: "zip",
				Files:  []*lockfile.File{{Name: "foo", Src: "foo"}},
				SLSAProvenance: &registry.SLSAProvenance{
					Type:  "github_release",
					Asset: new("foo-darwin-x64.intoto.jsonl"),
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			sel := asset.Select("foo/foo", "v1.0.0", d.sel, nil)
			// The last entry is the one under test: for Rosetta 2 it is the emulated arm64 one.
			chosen := sel.Assets[len(sel.Assets)-1]
			got, err := resolve.ResolveSelected(logger(), &resolve.Param{
				PkgName: "foo/foo",
				Version: "v1.0.0",
				PkgInfo: pkgInfo,
			}, chosen)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
