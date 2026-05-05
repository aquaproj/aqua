package aqua_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/google/go-cmp/cmp"
	"go.yaml.in/yaml/v2"
)

func TestConfig_UnmarshalYAML(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		yaml  string
		exp   *aqua.Config
	}{
		{
			title: "standard registry",
			yaml: `
registries:
- type: standard
  ref: v0.2.0
packages:
- name: cmdx
  registry: standard
  version: v1.6.0
`,
			exp: &aqua.Config{
				Registries: aqua.Registries{
					regTypeStandard: &aqua.Registry{
						Name:      regTypeStandard,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Path:      regFileRegistryYaml,
						Type:      pkgTypeGitHubContent,
						Ref:       "v0.2.0",
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "cmdx",
						Registry: regTypeStandard,
						Version:  "v1.6.0",
						Pin:      true,
					},
				},
			},
		},
		{
			title: "parse package name with version",
			yaml: `
registries:
- type: standard
  ref: v0.2.0
packages:
- name: suzuki-shunsuke/cmdx@v1.6.0
  registry: standard
`,
			exp: &aqua.Config{
				Registries: aqua.Registries{
					regTypeStandard: &aqua.Registry{
						Name:      regTypeStandard,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Path:      regFileRegistryYaml,
						Ref:       "v0.2.0",
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "suzuki-shunsuke/cmdx",
						Registry: regTypeStandard,
						Version:  "v1.6.0",
					},
				},
			},
		},
	}

	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			cfg := &aqua.Config{}
			if err := yaml.Unmarshal([]byte(d.yaml), cfg); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, cfg); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

// HasCommandAlias returns true if the given command alias is present.
func TestConfig_HasCommandAlias(t *testing.T) {
	t.Parallel()

	p := aqua.Package{
		CommandAliases: []*aqua.CommandAlias{
			{
				Command: pkgFoo,
				Alias:   "bar",
			},
		},
	}

	if p.HasCommandAlias(pkgFoo) {
		t.Fatal("HasCommandAlias(foo): wanted false, got true")
	}
	if !p.HasCommandAlias("bar") {
		t.Fatal("HasCommandAlias(bar): wanted true, got false")
	}
}
