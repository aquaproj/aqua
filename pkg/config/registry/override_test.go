package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/go-cmp/cmp"
)

func TestPackageInfo_Override(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		exp     *registry.PackageInfo
		isErr   bool
		pkgInfo *registry.PackageInfo
		version string
		rt      *runtime.Runtime
	}{
		{
			title: "not override",
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
				Replacements: registry.Replacements{
					"linux": "unknown-linux-musl",
				},
				Overrides: []*registry.Override{
					{
						GOOS:   "linux",
						GOArch: "arm64",
						Replacements: registry.Replacements{
							"linux": "unknown-linux-gnu",
						},
					},
				},
			},
			exp: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
				Replacements: registry.Replacements{
					"linux": "unknown-linux-musl",
				},
				Overrides: []*registry.Override{
					{
						GOOS:   "linux",
						GOArch: "arm64",
						Replacements: registry.Replacements{
							"linux": "unknown-linux-gnu",
						},
					},
				},
			},
			version: "v1.0.0",
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
		},
		{
			title: "override",
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
				Replacements: registry.Replacements{
					"linux": "unknown-linux-musl",
				},
				Overrides: []*registry.Override{
					{
						GOOS:   "linux",
						GOArch: "arm64",
						Replacements: registry.Replacements{
							"linux": "unknown-linux-gnu",
						},
					},
				},
			},
			exp: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
				Replacements: registry.Replacements{
					"linux": "unknown-linux-gnu",
				},
				Overrides: []*registry.Override{
					{
						GOOS:   "linux",
						GOArch: "arm64",
						Replacements: registry.Replacements{
							"linux": "unknown-linux-gnu",
						},
					},
				},
			},
			version: "v1.0.0",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "arm64",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo, err := d.pkgInfo.Override(d.version, d.rt)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
