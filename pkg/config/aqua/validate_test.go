//nolint:funlen
package aqua_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	data := []struct {
		name    string
		config  *aqua.Config
		wantErr bool
	}{
		{
			name: "valid config with standard registry",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeStandard: &aqua.Registry{
						Name:      regTypeStandard,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Ref:       versionV4,
						Path:      regFileRegistryYaml,
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "cli/cli",
						Version:  "v2.0.0",
						Registry: regTypeStandard,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with local registry",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeLocal: &aqua.Registry{
						Name: regTypeLocal,
						Type: regTypeLocal,
						Path: regFileRegistryYaml,
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "custom-tool",
						Version:  "v1.0.0",
						Registry: regTypeLocal,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid registry - missing repo_owner",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeInvalid: &aqua.Registry{
						Name:     regTypeInvalid,
						Type:     pkgTypeGitHubContent,
						RepoName: regNameAquaRegistry,
						Ref:      versionV4,
						Path:     regFileRegistryYaml,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry - missing repo_name",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeInvalid: &aqua.Registry{
						Name:      regTypeInvalid,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						Ref:       versionV4,
						Path:      regFileRegistryYaml,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry - missing ref",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeInvalid: &aqua.Registry{
						Name:      regTypeInvalid,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Path:      regFileRegistryYaml,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry - ref cannot be main",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeInvalid: &aqua.Registry{
						Name:      regTypeInvalid,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Ref:       "main",
						Path:      regFileRegistryYaml,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry - ref cannot be master",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeInvalid: &aqua.Registry{
						Name:      regTypeInvalid,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Ref:       "master",
						Path:      regFileRegistryYaml,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid local registry - missing path",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeInvalid: &aqua.Registry{
						Name: regTypeInvalid,
						Type: regTypeLocal,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry type",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeInvalid: &aqua.Registry{
						Name: regTypeInvalid,
						Type: "unknown",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty config",
			config: &aqua.Config{
				Registries: aqua.Registries{},
			},
			wantErr: false,
		},
		{
			name: "multiple valid registries",
			config: &aqua.Config{
				Registries: aqua.Registries{
					regTypeStandard: &aqua.Registry{
						Name:      regTypeStandard,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Ref:       versionV4,
						Path:      regFileRegistryYaml,
					},
					regTypeLocal: &aqua.Registry{
						Name: regTypeLocal,
						Type: regTypeLocal,
						Path: "local-registry.yaml",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple registries with one invalid",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"valid": &aqua.Registry{
						Name:      "valid",
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Ref:       versionV4,
						Path:      regFileRegistryYaml,
					},
					regTypeInvalid: &aqua.Registry{
						Name: regTypeInvalid,
						Type: pkgTypeGitHubContent,
						// Missing required fields
					},
				},
			},
			wantErr: true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			err := d.config.Validate()
			if d.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
