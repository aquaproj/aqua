//nolint:funlen
package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestCargo_YAML(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		yaml     string
		expected *registry.Cargo
	}{
		{
			name:     "empty cargo config",
			yaml:     `{}`,
			expected: &registry.Cargo{},
		},
		{
			name: "cargo with features",
			yaml: `
features:
  - default
  - tls
`,
			expected: &registry.Cargo{
				Features: []string{"default", "tls"},
			},
		},
		{
			name: "cargo with all features enabled",
			yaml: `
all_features: true
`,
			expected: &registry.Cargo{
				AllFeatures: true,
			},
		},
		{
			name: "cargo with locked enabled",
			yaml: `
locked: true
`,
			expected: &registry.Cargo{
				Locked: true,
			},
		},
		{
			name: "full cargo configuration",
			yaml: `
features:
  - cli
  - json
all_features: false
locked: true
`,
			expected: &registry.Cargo{
				Features:    []string{"cli", "json"},
				AllFeatures: false,
				Locked:      true,
			},
		},
		{
			name: "cargo with single feature",
			yaml: `
features: ["json"]
`,
			expected: &registry.Cargo{
				Features: []string{"json"},
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			var cargo registry.Cargo
			if err := yaml.Unmarshal([]byte(d.yaml), &cargo); err != nil {
				t.Fatalf("failed to unmarshal YAML: %v", err)
			}

			if diff := cmp.Diff(d.expected, &cargo); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
