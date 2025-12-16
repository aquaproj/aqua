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
						Registries: []*policy.Registry{
							{
								Type:      "github_content",
								Name:      registryTypeStandard,
								RepoOwner: "aquaproj",
								RepoName:  "aqua",
								Path:      "registry.yaml",
							},
						},
						Packages: []*policy.Package{
							{
								Name:         "cli/cli",
								RegistryName: registryTypeStandard,
							},
							{
								Name:         "suzuki-shunsuke/tfcmt",
								Version:      `semver(">= 3.0.0")`,
								RegistryName: "standard",
							},
						},
					},
				},
			},
		},
		{
			name: "http registry allowed",
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    "suzuki-shunsuke/ghalint",
					Version: "v1.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &aqua.Registry{
					Type:    "http",
					Name:    "http-raw",
					URL:     "http://localhost:8888/v{{.Version}}/registry.yaml",
					Version: "1.0.0",
					Format:  "raw",
				},
			},
			policies: []*policy.Config{
				{
					YAML: &policy.ConfigYAML{
						Registries: []*policy.Registry{
							{
								Type: "http",
								Name: "http-raw",
								URL:  "http://localhost:8888/v{{.Version}}/registry.yaml",
								// No version specified means any version is allowed
							},
						},
						Packages: []*policy.Package{
							{
								RegistryName: "http-raw",
							},
						},
					},
				},
			},
		},
		{
			name:  "http registry not allowed - wrong URL",
			isErr: true,
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    "suzuki-shunsuke/ghalint",
					Version: "v1.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &aqua.Registry{
					Type:    "http",
					Name:    "http-raw",
					URL:     "http://malicious.com/v{{.Version}}/registry.yaml",
					Version: "1.0.0",
					Format:  "raw",
				},
			},
			policies: []*policy.Config{
				{
					YAML: &policy.ConfigYAML{
						Registries: []*policy.Registry{
							{
								Type:    "http",
								Name:    "http-raw",
								URL:     "http://localhost:8888/v{{.Version}}/registry.yaml",
								Version: "1.0.0",
							},
						},
						Packages: []*policy.Package{
							{
								RegistryName: "http-raw",
							},
						},
					},
				},
			},
		},
		{
			name: "http registry with version constraint",
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    "suzuki-shunsuke/ghalint",
					Version: "v1.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &aqua.Registry{
					Type:    "http",
					Name:    "http-raw",
					URL:     "http://localhost:8888/v{{.Version}}/registry.yaml",
					Version: "1.5.0",
					Format:  "raw",
				},
			},
			policies: []*policy.Config{
				{
					YAML: &policy.ConfigYAML{
						Registries: []*policy.Registry{
							{
								Type:    "http",
								Name:    "http-raw",
								URL:     "http://localhost:8888/v{{.Version}}/registry.yaml",
								Version: `semver(">= 1.0.0")`,
							},
						},
						Packages: []*policy.Package{
							{
								RegistryName: "http-raw",
							},
						},
					},
				},
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			// Initialize policies to populate registry mappings
			for _, pol := range d.policies {
				if err := pol.Init(); err != nil {
					t.Fatal(err)
				}
			}
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
