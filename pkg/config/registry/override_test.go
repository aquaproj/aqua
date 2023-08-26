package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
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
				Asset:     ptr.String("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
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
				Asset:     ptr.String("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
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
				Asset:     ptr.String("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
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
				Asset:     ptr.String("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
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
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo, err := d.pkgInfo.Override(logE, d.version, d.rt)
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

func TestOverride_Match(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exp      bool
		override *registry.Override
		rt       *runtime.Runtime
	}{
		{
			title: "goos doesn't match",
			override: &registry.Override{
				GOOS: "linux",
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
		},
		{
			title: "goarch doesn't match",
			override: &registry.Override{
				GOArch: "arm64",
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
		},
		{
			title: "match",
			exp:   true,
			override: &registry.Override{
				GOOS: "darwin",
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if f := d.override.Match(d.rt); f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
