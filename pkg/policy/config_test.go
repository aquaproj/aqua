package policy_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/policy"
	"github.com/google/go-cmp/cmp"
)

const (
	registryTypeStandard = "standard"
	registryTypeLocal    = "local"
)

func TestConfig_Init(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name  string
		cfg   *policy.Config
		isErr bool
		exp   *policy.Config
	}{
		{
			name: "normal",
			cfg: &policy.Config{
				Path: "/home/foo/aqua-policy.yaml",
				YAML: &policy.ConfigYAML{
					Registries: []*policy.Registry{
						{
							Type: registryTypeStandard,
						},
						{
							Type: registryTypeLocal,
							Path: "registry.yaml",
							Name: "foo",
						},
					},
					Packages: []*policy.Package{
						{},
						{
							RegistryName: "foo",
						},
					},
				},
			},
			exp: &policy.Config{
				Path: "/home/foo/aqua-policy.yaml",
				YAML: &policy.ConfigYAML{
					Registries: []*policy.Registry{
						{
							Type:      "github_content",
							Name:      registryTypeStandard,
							RepoOwner: "aquaproj",
							RepoName:  "aqua",
							Path:      "registry.yaml",
						},
						{
							Type: registryTypeLocal,
							Path: "/home/foo/registry.yaml",
							Name: "foo",
						},
					},
					Packages: []*policy.Package{
						{
							RegistryName: registryTypeStandard,
							Registry: &policy.Registry{
								Type:      "github_content",
								Name:      registryTypeStandard,
								RepoOwner: "aquaproj",
								RepoName:  "aqua",
								Path:      "registry.yaml",
							},
						},
						{
							RegistryName: "foo",
							Registry: &policy.Registry{
								Type: registryTypeLocal,
								Path: "/home/foo/registry.yaml",
								Name: "foo",
							},
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if err := d.cfg.Init(); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
				if diff := cmp.Diff(d.exp, d.cfg); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
