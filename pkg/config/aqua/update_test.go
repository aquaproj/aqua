//nolint:funlen
package aqua_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

func TestUpdate_GetEnabled(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		update   *aqua.Update
		expected bool
	}{
		{
			name:     "update is nil",
			update:   nil,
			expected: true,
		},
		{
			name:     "enabled is nil (default true)",
			update:   &aqua.Update{},
			expected: true,
		},
		{
			name: "enabled is true",
			update: &aqua.Update{
				Enabled: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "enabled is false",
			update: &aqua.Update{
				Enabled: boolPtr(false),
			},
			expected: false,
		},
		{
			name: "full update configuration with enabled true",
			update: &aqua.Update{
				Enabled:        boolPtr(true),
				AllowedVersion: "semver('>=1.0.0')",
				Types:          []string{"major", "minor", "patch"},
			},
			expected: true,
		},
		{
			name: "full update configuration with enabled false",
			update: &aqua.Update{
				Enabled:        boolPtr(false),
				AllowedVersion: "semver('>=1.0.0')",
				Types:          []string{"minor", "patch"},
			},
			expected: false,
		},
		{
			name: "allowed version with nil enabled",
			update: &aqua.Update{
				AllowedVersion: "Version != '1.0.0'",
			},
			expected: true,
		},
		{
			name: "types constraint with nil enabled",
			update: &aqua.Update{
				Types: []string{"patch"},
			},
			expected: true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.update.GetEnabled()
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

// boolPtr helper function is defined in checksum_test.go
