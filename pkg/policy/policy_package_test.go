package policy_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/policy"
)

func TestChecker_ValidatePackage(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name  string
		isErr bool
		param *policy.ParamValidatePackage
	}{
		{
			name: "no policy",
			param: &policy.ParamValidatePackage{
				Pkg: &config.Package{},
			},
		},
		{
			name: "normal",
			param: &policy.ParamValidatePackage{
				Pkg: &config.Package{
					Package: &aqua.Package{
						Name:    "suzuki-shunsuke/tfcmt",
						Version: "v4.0.0",
					},
					Registry: &aqua.Registry{
						Type:      "github_content",
						Name:      registryTypeStandard,
						RepoOwner: "aquaproj",
						RepoName:  "aqua",
						Path:      "registry.yaml",
						Ref:       "v1.90.0",
					},
				},
				PolicyConfigs: []*policy.Config{
					{
						YAML: &policy.ConfigYAML{
							Packages: []*policy.Package{
								{
									Name: "cli/cli",
								},
								{
									Name:         "suzuki-shunsuke/tfcmt",
									Version:      `semver(">= 3.0.0")`,
									RegistryName: "standard",
									Registry: &policy.Registry{
										Type:      "github_content",
										Name:      registryTypeStandard,
										RepoOwner: "aquaproj",
										RepoName:  "aqua",
										Path:      "registry.yaml",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	checker := &policy.Checker{}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if err := checker.ValidatePackage(d.param); err != nil {
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
