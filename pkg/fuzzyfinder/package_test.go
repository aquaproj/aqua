package fuzzyfinder_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
)

const registryStandard = "standard"

func TestPackage_Item(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name string
		pkg  *fuzzyfinder.Package
		exp  string
	}{
		{
			name: "normal",
			pkg: &fuzzyfinder.Package{
				PackageInfo: &registry.PackageInfo{
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
				},
				RegistryName: registryStandard,
			},
			exp: "suzuki-shunsuke/ci-info",
		},
		{
			name: "search words",
			pkg: &fuzzyfinder.Package{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "suzuki-shunsuke",
					RepoName:    "ci-info",
					SearchWords: []string{"pull request"},
				},
				RegistryName: registryStandard,
			},
			exp: "suzuki-shunsuke/ci-info: pull request",
		},
		{
			name: "search words, registry",
			pkg: &fuzzyfinder.Package{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "suzuki-shunsuke",
					RepoName:    "ci-info",
					SearchWords: []string{"pull request"},
				},
				RegistryName: "foo",
			},
			exp: "suzuki-shunsuke/ci-info (foo): pull request",
		},
		{
			name: "search words, alias, registry",
			pkg: &fuzzyfinder.Package{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "suzuki-shunsuke",
					RepoName:    "ci-info",
					SearchWords: []string{"pull request"},
					Aliases: []*registry.Alias{
						{
							Name: "ci-info",
						},
					},
				},
				RegistryName: "foo",
			},
			exp: "suzuki-shunsuke/ci-info (ci-info) (foo): pull request",
		},
		{
			name: "search words, alias, command, registry",
			pkg: &fuzzyfinder.Package{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "suzuki-shunsuke",
					RepoName:    "ci-info",
					SearchWords: []string{"pull request"},
					Aliases: []*registry.Alias{
						{
							Name: "ci-info",
						},
					},
					Files: []*registry.File{
						{
							Name: "ci-info",
						},
						{
							Name: "ci",
						},
					},
				},
				RegistryName: "foo",
			},
			exp: "suzuki-shunsuke/ci-info (ci-info) (foo) [ci-info, ci]: pull request",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			s := d.pkg.Item()
			if s != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}
