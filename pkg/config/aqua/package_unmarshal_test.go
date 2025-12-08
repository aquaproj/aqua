//nolint:funlen
package aqua_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"go.yaml.in/yaml/v2"
)

func TestPackage_UnmarshalYAML_ParseNameWithVersion(t *testing.T) {
	t.Parallel()
	data := []struct {
		name            string
		yamlContent     string
		expectedName    string
		expectedVersion string
		expectedPin     bool
	}{
		{
			name: "name with version",
			yamlContent: `
name: cli/cli@v2.0.0
`,
			expectedName:    "cli/cli",
			expectedVersion: "v2.0.0",
			expectedPin:     false, // Version from name doesn't set pin
		},
		{
			name: "name without version",
			yamlContent: `
name: cli/cli
`,
			expectedName:    "cli/cli",
			expectedVersion: "",
			expectedPin:     false,
		},
		{
			name: "name with multiple @ symbols",
			yamlContent: `
name: scope/package@1.0.0@beta
`,
			expectedName:    "scope/package",
			expectedVersion: "1.0.0@beta",
			expectedPin:     false,
		},
		{
			name: "only version in name",
			yamlContent: `
name: "@v1.0.0"
`,
			expectedName:    "@v1.0.0", // Original name preserved when parsed name is empty
			expectedVersion: "v1.0.0",  // Version after @
			expectedPin:     false,
		},
		{
			name: "name ending with @",
			yamlContent: `
name: package@
`,
			expectedName:    "package",
			expectedVersion: "",
			expectedPin:     false,
		},
		{
			name: "complex package name with version",
			yamlContent: `
name: github.com/user/repo@v1.2.3-alpha
`,
			expectedName:    "github.com/user/repo",
			expectedVersion: "v1.2.3-alpha",
			expectedPin:     false,
		},
		{
			name: "both name@version and explicit version",
			yamlContent: `
name: cli/cli@v1.0.0
version: v2.0.0
`,
			expectedName:    "cli/cli",
			expectedVersion: "v1.0.0", // Name@version overwrites explicit version
			expectedPin:     true,     // Explicit version sets pin
		},
		{
			name: "explicit version only",
			yamlContent: `
name: cli/cli
version: v2.0.0
`,
			expectedName:    "cli/cli",
			expectedVersion: "v2.0.0",
			expectedPin:     true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			var pkg aqua.Package
			if err := yaml.Unmarshal([]byte(d.yamlContent), &pkg); err != nil {
				t.Fatalf("UnmarshalYAML failed: %v", err)
			}

			if pkg.Name != d.expectedName {
				t.Errorf("expected name %q, got %q", d.expectedName, pkg.Name)
			}
			if pkg.Version != d.expectedVersion {
				t.Errorf("expected version %q, got %q", d.expectedVersion, pkg.Version)
			}
			if pkg.Pin != d.expectedPin {
				t.Errorf("expected pin %v, got %v", d.expectedPin, pkg.Pin)
			}
			// Default registry should be set
			if pkg.Registry != "standard" {
				t.Errorf("expected registry %q, got %q", "standard", pkg.Registry)
			}
		})
	}
}
