package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/go-cmp/cmp"
)

func TestPackageInfo_Override(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		exp     *config.PackageInfo
		isErr   bool
		pkgInfo *config.PackageInfo
		version string
		rt      *runtime.Runtime
	}{
		{
			title: "not override",
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
				Replacements: map[string]string{
					"linux": "unknown-linux-musl",
				},
				Overrides: []*config.Override{
					{
						GOOS:   "linux",
						GOArch: "arm64",
						Replacements: map[string]string{
							"linux": "unknown-linux-gnu",
						},
					},
				},
			},
			exp: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
				Replacements: map[string]string{
					"linux": "unknown-linux-musl",
				},
				Overrides: []*config.Override{
					{
						GOOS:   "linux",
						GOArch: "arm64",
						Replacements: map[string]string{
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
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
				Replacements: map[string]string{
					"linux": "unknown-linux-musl",
				},
				Overrides: []*config.Override{
					{
						GOOS:   "linux",
						GOArch: "arm64",
						Replacements: map[string]string{
							"linux": "unknown-linux-gnu",
						},
					},
				},
			},
			exp: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
				Replacements: map[string]string{
					"linux": "unknown-linux-gnu",
				},
				Overrides: []*config.Override{
					{
						GOOS:   "linux",
						GOArch: "arm64",
						Replacements: map[string]string{
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
