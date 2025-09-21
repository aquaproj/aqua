//nolint:funlen
package aqua_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestRegistry_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		yaml     string
		expected *aqua.Registry
	}{
		{
			name: "standard registry with defaults",
			yaml: `
type: standard
ref: v4.0.0
`,
			expected: &aqua.Registry{
				Name:      "standard",
				Type:      "github_content",
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Path:      "registry.yaml",
				Ref:       "v4.0.0",
			},
		},
		{
			name: "standard registry with custom name",
			yaml: `
name: my-standard
type: standard
ref: v3.0.0
`,
			expected: &aqua.Registry{
				Name:      "my-standard",
				Type:      "github_content",
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Path:      "registry.yaml",
				Ref:       "v3.0.0",
			},
		},
		{
			name: "standard registry with custom repo_owner",
			yaml: `
type: standard
repo_owner: custom-org
ref: v4.0.0
`,
			expected: &aqua.Registry{
				Name:      "standard",
				Type:      "github_content",
				RepoOwner: "custom-org",
				RepoName:  "aqua-registry",
				Path:      "registry.yaml",
				Ref:       "v4.0.0",
			},
		},
		{
			name: "standard registry with custom repo_name",
			yaml: `
type: standard
repo_name: custom-registry
ref: v4.0.0
`,
			expected: &aqua.Registry{
				Name:      "standard",
				Type:      "github_content",
				RepoOwner: "aquaproj",
				RepoName:  "custom-registry",
				Path:      "registry.yaml",
				Ref:       "v4.0.0",
			},
		},
		{
			name: "standard registry with custom path",
			yaml: `
type: standard
path: custom.yaml
ref: v4.0.0
`,
			expected: &aqua.Registry{
				Name:      "standard",
				Type:      "github_content",
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Path:      "custom.yaml",
				Ref:       "v4.0.0",
			},
		},
		{
			name: "github_content registry",
			yaml: `
name: custom
type: github_content
repo_owner: example
repo_name: example-registry
path: registry.yaml
ref: v1.0.0
`,
			expected: &aqua.Registry{
				Name:      "custom",
				Type:      "github_content",
				RepoOwner: "example",
				RepoName:  "example-registry",
				Path:      "registry.yaml",
				Ref:       "v1.0.0",
			},
		},
		{
			name: "local registry",
			yaml: `
name: local
type: local
path: ./local-registry.yaml
`,
			expected: &aqua.Registry{
				Name: "local",
				Type: "local",
				Path: "./local-registry.yaml",
			},
		},
		{
			name: "registry with private flag",
			yaml: `
name: private-reg
type: github_content
repo_owner: private-org
repo_name: private-registry
path: registry.yaml
ref: v1.0.0
private: true
`,
			expected: &aqua.Registry{
				Name:      "private-reg",
				Type:      "github_content",
				RepoOwner: "private-org",
				RepoName:  "private-registry",
				Path:      "registry.yaml",
				Ref:       "v1.0.0",
				Private:   true,
			},
		},
		{
			name: "standard registry with all custom fields",
			yaml: `
name: custom-standard
type: standard
repo_owner: custom-org
repo_name: custom-registry
path: custom.yaml
ref: v2.0.0
private: true
`,
			expected: &aqua.Registry{
				Name:      "custom-standard",
				Type:      "github_content",
				RepoOwner: "custom-org",
				RepoName:  "custom-registry",
				Path:      "custom.yaml",
				Ref:       "v2.0.0",
				Private:   true,
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			var registry aqua.Registry
			if err := yaml.Unmarshal([]byte(d.yaml), &registry); err != nil {
				t.Fatalf("failed to unmarshal YAML: %v", err)
			}

			if diff := cmp.Diff(d.expected, &registry); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRegistries_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		yaml     string
		expected aqua.Registries
	}{
		{
			name: "single registry",
			yaml: `
- name: standard
  type: standard
  ref: v4.0.0
`,
			expected: aqua.Registries{
				"standard": &aqua.Registry{
					Name:      "standard",
					Type:      "github_content",
					RepoOwner: "aquaproj",
					RepoName:  "aqua-registry",
					Path:      "registry.yaml",
					Ref:       "v4.0.0",
				},
			},
		},
		{
			name: "multiple registries",
			yaml: `
- name: standard
  type: standard
  ref: v4.0.0
- name: local
  type: local
  path: ./local.yaml
`,
			expected: aqua.Registries{
				"standard": &aqua.Registry{
					Name:      "standard",
					Type:      "github_content",
					RepoOwner: "aquaproj",
					RepoName:  "aqua-registry",
					Path:      "registry.yaml",
					Ref:       "v4.0.0",
				},
				"local": &aqua.Registry{
					Name: "local",
					Type: "local",
					Path: "./local.yaml",
				},
			},
		},
		{
			name:     "empty registries",
			yaml:     `[]`,
			expected: aqua.Registries{},
		},
		{
			name: "registry with nil entry (should be ignored)",
			yaml: `
- name: standard
  type: standard
  ref: v4.0.0
- null
`,
			expected: aqua.Registries{
				"standard": &aqua.Registry{
					Name:      "standard",
					Type:      "github_content",
					RepoOwner: "aquaproj",
					RepoName:  "aqua-registry",
					Path:      "registry.yaml",
					Ref:       "v4.0.0",
				},
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			var registries aqua.Registries
			if err := yaml.Unmarshal([]byte(d.yaml), &registries); err != nil {
				t.Fatalf("failed to unmarshal YAML: %v", err)
			}

			if diff := cmp.Diff(d.expected, registries); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
