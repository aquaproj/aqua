package controller_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/aqua/pkg/controller"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYAML(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		yaml  string
		exp   *controller.Config
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
			exp: &controller.Config{
				Registries: controller.Registries{
					&controller.GitHubContentRegistry{
						Name:      "standard",
						RepoOwner: "suzuki-shunsuke",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
						Ref:       "v0.2.0",
					},
				},
				Packages: []*controller.Package{
					{
						Name:     "cmdx",
						Registry: "standard",
						Version:  "v1.6.0",
					},
				},
			},
		},
		{
			title: "inline registry",
			yaml: `
inline_registry:
  packages:
  - name: cmdx
    type: github_release
    repo_owner: suzuki-shunsuke
    repo_name: cmdx
    asset: 'cmdx_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz'
    files:
    - name: cmdx
packages:
- name: cmdx
  registry: inline
  version: v1.6.0
`,
			exp: &controller.Config{
				InlineRegistry: &controller.RegistryContent{
					PackageInfos: controller.PackageInfos{
						&controller.GitHubReleasePackageInfo{
							Name:      "cmdx",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "cmdx",
							Asset:     controller.NewTemplate(`cmdx_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz`),
							Files: []*controller.File{
								{
									Name: "cmdx",
								},
							},
						},
					},
				},
				Packages: []*controller.Package{
					{
						Name:     "cmdx",
						Registry: "inline",
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
			exp: &controller.Config{
				Registries: controller.Registries{
					&controller.GitHubContentRegistry{
						Name:      "standard",
						RepoOwner: "suzuki-shunsuke",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
						Ref:       "v0.2.0",
					},
				},
				Packages: []*controller.Package{
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
			cfg := &controller.Config{}
			if err := yaml.Unmarshal([]byte(d.yaml), cfg); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, cfg, cmp.AllowUnexported(controller.Template{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPackageInfos_ToMap(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		pkgInfos *controller.PackageInfos
		exp      map[string]controller.PackageInfo
		isErr    bool
	}{
		{
			title: "normal",
			pkgInfos: &controller.PackageInfos{
				&controller.HTTPPackageInfo{
					Name: "foo",
				},
			},
			exp: map[string]controller.PackageInfo{
				"foo": &controller.HTTPPackageInfo{
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
