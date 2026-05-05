package aqua_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

func TestFilterPackageByTag(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		pkg          *aqua.Package
		tags         map[string]struct{}
		excludedTags map[string]struct{}
		exp          bool
	}{
		{
			name: "no tag",
			pkg: &aqua.Package{
				Name:     repoCliCli,
				Version:  versionV2,
				Registry: regTypeStandard,
			},
			exp: true,
		},
		{
			name: "package has tags but no tag is specified",
			pkg: &aqua.Package{
				Name:     repoCliCli,
				Version:  versionV2,
				Registry: regTypeStandard,
				Tags:     []string{"ci"},
			},
			exp: true,
		},
		{
			name: "tag is matched",
			pkg: &aqua.Package{
				Name:     repoCliCli,
				Version:  versionV2,
				Registry: regTypeStandard,
				Tags:     []string{"ci", pkgFoo},
			},
			tags: map[string]struct{}{
				"ci": {},
			},
			exp: true,
		},
		{
			name: "exclude tag is matched",
			pkg: &aqua.Package{
				Name:     repoCliCli,
				Version:  versionV2,
				Registry: regTypeStandard,
				Tags:     []string{"ci", pkgFoo},
			},
			excludedTags: map[string]struct{}{
				"ci": {},
			},
			exp: false,
		},
		{
			name: "exclude tag and tag are matched",
			pkg: &aqua.Package{
				Name:     repoCliCli,
				Version:  versionV2,
				Registry: regTypeStandard,
				Tags:     []string{"ci", pkgFoo},
			},
			tags: map[string]struct{}{
				pkgFoo: {},
			},
			excludedTags: map[string]struct{}{
				"ci": {},
			},
			exp: false,
		},
		{
			name: "exclude tag isn't matched and tag is matched",
			pkg: &aqua.Package{
				Name:     repoCliCli,
				Version:  versionV2,
				Registry: regTypeStandard,
				Tags:     []string{"ci", pkgFoo},
			},
			tags: map[string]struct{}{
				pkgFoo: {},
			},
			excludedTags: map[string]struct{}{
				pkgYoo: {},
			},
			exp: true,
		},
		{
			name: "exclude tag and tag aren't matched",
			pkg: &aqua.Package{
				Name:     repoCliCli,
				Version:  versionV2,
				Registry: regTypeStandard,
				Tags:     []string{"ci"},
			},
			tags: map[string]struct{}{
				pkgFoo: {},
			},
			excludedTags: map[string]struct{}{
				pkgYoo: {},
			},
			exp: false,
		},
		{
			name: "exclude tag ins't matched and tag isn't specified",
			pkg: &aqua.Package{
				Name:     repoCliCli,
				Version:  versionV2,
				Registry: regTypeStandard,
				Tags:     []string{"ci"},
			},
			excludedTags: map[string]struct{}{
				pkgYoo: {},
			},
			exp: true,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			f := aqua.FilterPackageByTag(d.pkg, d.tags, d.excludedTags)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
