package generate

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
)

func Test_find(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		pkg  *FindingPackage
		exp  string
	}{
		{
			name: "normal",
			pkg: &FindingPackage{
				PackageInfo: &config.PackageInfo{
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
				},
				RegistryName: "standard",
			},
			exp: "suzuki-shunsuke/ci-info (standard)",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			s := find(d.pkg)
			if s != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}

func Test_getPreview(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		pkg  *FindingPackage
		i    int
		w    int
		exp  string
	}{
		{
			name: "normal",
			pkg: &FindingPackage{
				PackageInfo: &config.PackageInfo{
					RepoOwner:   "suzuki-shunsuke",
					RepoName:    "ci-info",
					Description: "CLI tool to get CI related information",
				},
				RegistryName: "standard",
			},
			w: 100,
			exp: `suzuki-shunsuke/ci-info

https://github.com/suzuki-shunsuke/ci-info
CLI tool to get CI related information`,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			s := getPreview(d.pkg, d.i, d.w)
			if s != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}
