package policy_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
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
					Name:    "suzuki-shunsuke/tfcmt",
					Version: "v4.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &aqua.Registry{
					Type:      "github_content",
					Name:      registryTypeStandard,
					RepoOwner: "aquaproj",
					RepoName:  "aqua-registry",
					Path:      "registry.yaml",
					Ref:       "v3.90.0",
				},
			},
		},
		{
			name: "normal",
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    "suzuki-shunsuke/tfcmt",
					Version: "v4.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &aqua.Registry{
					Type:      "github_content",
					Name:      registryTypeStandard,
					RepoOwner: "aquaproj",
					RepoName:  "aqua",
					Path:      "registry.yaml",
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
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if err := policy.ValidatePackage(logE, d.pkg, d.policies); err != nil {
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
