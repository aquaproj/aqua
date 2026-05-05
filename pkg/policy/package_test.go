package policy_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/policy"
)

func TestValidatePackage(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		isErr    bool
		pkg      *config.Package
		policies []*policy.Config
	}{
		{
			name: "no policy",
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    repoSuzukiTfcmt,
					Version: "v4.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &aqua.Registry{
					Type:      pkgTypeGitHubContent,
					Name:      registryTypeStandard,
					RepoOwner: regOwnerAquaproj,
					RepoName:  regNameAquaRegistry,
					Path:      regFileRegistryYaml,
					Ref:       "v3.90.0",
				},
			},
		},
		{
			name: caseNormal,
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    repoSuzukiTfcmt,
					Version: "v4.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &aqua.Registry{
					Type:      pkgTypeGitHubContent,
					Name:      registryTypeStandard,
					RepoOwner: regOwnerAquaproj,
					RepoName:  "aqua",
					Path:      regFileRegistryYaml,
					Ref:       "v1.90.0",
				},
			},
			policies: []*policy.Config{
				{
					YAML: &policy.ConfigYAML{
						Packages: []*policy.Package{
							{
								Name: "cli/cli",
							},
							{
								Name:         repoSuzukiTfcmt,
								Version:      `semver(">= 3.0.0")`,
								RegistryName: "standard",
								Registry: &policy.Registry{
									Type:      pkgTypeGitHubContent,
									Name:      registryTypeStandard,
									RepoOwner: regOwnerAquaproj,
									RepoName:  "aqua",
									Path:      regFileRegistryYaml,
								},
							},
						},
					},
				},
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if err := policy.ValidatePackage(logger, d.pkg, d.policies); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
