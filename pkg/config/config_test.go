package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYAML(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		yaml  string
		exp   *config.Config
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
			exp: &config.Config{
				Registries: config.Registries{
					"standard": &config.Registry{
						Name:      "standard",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
						Type:      "github_content",
						Ref:       "v0.2.0",
					},
				},
				Packages: []*config.Package{
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
			exp: &config.Config{
				Registries: config.Registries{
					"standard": &config.Registry{
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
						Ref:       "v0.2.0",
					},
				},
				Packages: []*config.Package{
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
			cfg := &config.Config{}
			if err := yaml.Unmarshal([]byte(d.yaml), cfg); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, cfg); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPackageInfos_ToMap(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		pkgInfos *config.PackageInfos
		exp      map[string]*config.PackageInfo
		isErr    bool
	}{
		{
			title: "normal",
			pkgInfos: &config.PackageInfos{
				&config.PackageInfo{
					Type: "github_release",
					Name: "foo",
				},
			},
			exp: map[string]*config.PackageInfo{
				"foo": {
					Type: "github_release",
					Name: "foo",
				},
			},
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			m, err := d.pkgInfos.ToMap()
			if d.isErr {
				if err == nil {
					t.Fatal("error should be returned")
				}
				return
			}
			if diff := cmp.Diff(d.exp, m); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
