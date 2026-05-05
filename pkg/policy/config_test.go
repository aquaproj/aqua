package policy_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/policy"
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
			name: caseNormal,
			cfg: &policy.Config{
				Path: "/home/foo/aqua-policy.yaml",
				YAML: &policy.ConfigYAML{
					Registries: []*policy.Registry{
						{
							Type: registryTypeStandard,
						},
						{
							Type: registryTypeLocal,
							Path: regFileRegistryYaml,
							Name: pkgFoo,
						},
					},
					Packages: []*policy.Package{
						{},
						{
							RegistryName: pkgFoo,
						},
					},
				},
			},
			exp: &policy.Config{
				Path: "/home/foo/aqua-policy.yaml",
				YAML: &policy.ConfigYAML{
					Registries: []*policy.Registry{
						{
							Type:      pkgTypeGitHubContent,
							Name:      registryTypeStandard,
							RepoOwner: regOwnerAquaproj,
							RepoName:  regNameAquaRegistry,
							Path:      regFileRegistryYaml,
						},
						{
							Type: registryTypeLocal,
							Path: "/home/foo/registry.yaml",
							Name: pkgFoo,
						},
					},
					Packages: []*policy.Package{
						{
							RegistryName: registryTypeStandard,
							Registry: &policy.Registry{
								Type:      pkgTypeGitHubContent,
								Name:      registryTypeStandard,
								RepoOwner: regOwnerAquaproj,
								RepoName:  regNameAquaRegistry,
								Path:      regFileRegistryYaml,
							},
						},
						{
							RegistryName: pkgFoo,
							Registry: &policy.Registry{
								Type: registryTypeLocal,
								Path: "/home/foo/registry.yaml",
								Name: pkgFoo,
							},
						},
					},
				},
			},
		},
	}
	for _, d := range data {
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
			}
			if diff := cmp.Diff(d.exp, d.cfg); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
