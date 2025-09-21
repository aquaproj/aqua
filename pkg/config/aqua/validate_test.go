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
					"standard": &aqua.Registry{
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v4.0.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "cli/cli",
						Version:  "v2.0.0",
						Registry: "standard",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with local registry",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"local": &aqua.Registry{
						Name: "local",
						Type: "local",
						Path: "registry.yaml",
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "custom-tool",
						Version:  "v1.0.0",
						Registry: "local",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid registry - missing repo_owner",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"invalid": &aqua.Registry{
						Name:     "invalid",
						Type:     "github_content",
						RepoName: "aqua-registry",
						Ref:      "v4.0.0",
						Path:     "registry.yaml",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry - missing repo_name",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"invalid": &aqua.Registry{
						Name:      "invalid",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						Ref:       "v4.0.0",
						Path:      "registry.yaml",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry - missing ref",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"invalid": &aqua.Registry{
						Name:      "invalid",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry - ref cannot be main",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"invalid": &aqua.Registry{
						Name:      "invalid",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "main",
						Path:      "registry.yaml",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry - ref cannot be master",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"invalid": &aqua.Registry{
						Name:      "invalid",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "master",
						Path:      "registry.yaml",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid local registry - missing path",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"invalid": &aqua.Registry{
						Name: "invalid",
						Type: "local",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid registry type",
			config: &aqua.Config{
				Registries: aqua.Registries{
					"invalid": &aqua.Registry{
						Name: "invalid",
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
					"standard": &aqua.Registry{
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v4.0.0",
						Path:      "registry.yaml",
					},
					"local": &aqua.Registry{
						Name: "local",
						Type: "local",
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
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v4.0.0",
						Path:      "registry.yaml",
					},
					"invalid": &aqua.Registry{
						Name: "invalid",
						Type: "github_content",
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
