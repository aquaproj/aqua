package aqua_test

import (
	"testing"

	"github.com/clivm/clivm/pkg/config/aqua"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYAML(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		yaml  string
		exp   *clivm.Config
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
			exp: &clivm.Config{
				Registries: clivm.Registries{
					"standard": &clivm.Registry{
						Name:      "standard",
						RepoOwner: "clivm",
						RepoName:  "clivm-registry",
						Path:      "registry.yaml",
						Type:      "github_content",
						Ref:       "v0.2.0",
					},
				},
				Packages: []*clivm.Package{
					{
						Name:     "cmdx",
						Registry: "standard",
						Version:  "v1.6.0",
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
			exp: &clivm.Config{
				Registries: clivm.Registries{
					"standard": &clivm.Registry{
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "clivm",
						RepoName:  "clivm-registry",
						Path:      "registry.yaml",
						Ref:       "v0.2.0",
					},
				},
				Packages: []*clivm.Package{
					{
						Name:     "suzuki-shunsuke/cmdx",
						Registry: "standard",
						Version:  "v1.6.0",
					},
				},
			},
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			cfg := &clivm.Config{}
			if err := yaml.Unmarshal([]byte(d.yaml), cfg); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, cfg); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
