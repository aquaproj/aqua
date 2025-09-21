//nolint:funlen
package aqua_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

func TestConfig_ChecksumEnabled(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name         string
		config       *aqua.Config
		enforceValue bool
		defValue     bool
		expected     bool
	}{
		{
			name:         "enforce value is true",
			config:       &aqua.Config{},
			enforceValue: true,
			defValue:     false,
			expected:     true,
		},
		{
			name:         "config is nil, use default",
			config:       nil,
			enforceValue: false,
			defValue:     true,
			expected:     true,
		},
		{
			name:         "checksum is nil, use default",
			config:       &aqua.Config{},
			enforceValue: false,
			defValue:     false,
			expected:     false,
		},
		{
			name: "checksum enabled is nil, use default",
			config: &aqua.Config{
				Checksum: &aqua.Checksum{},
			},
			enforceValue: false,
			defValue:     true,
			expected:     true,
		},
		{
			name: "checksum enabled is true",
			config: &aqua.Config{
				Checksum: &aqua.Checksum{
					Enabled: boolPtr(true),
				},
			},
			enforceValue: false,
			defValue:     false,
			expected:     true,
		},
		{
			name: "checksum enabled is false",
			config: &aqua.Config{
				Checksum: &aqua.Checksum{
					Enabled: boolPtr(false),
				},
			},
			enforceValue: false,
			defValue:     true,
			expected:     false,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.config.ChecksumEnabled(d.enforceValue, d.defValue)
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

func TestConfig_RequireChecksum(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name         string
		config       *aqua.Config
		enforceValue bool
		defValue     bool
		expected     bool
	}{
		{
			name:         "enforce value is true",
			config:       &aqua.Config{},
			enforceValue: true,
			defValue:     false,
			expected:     true,
		},
		{
			name:         "config is nil, use default",
			config:       nil,
			enforceValue: false,
			defValue:     true,
			expected:     true,
		},
		{
			name:         "checksum is nil, use default",
			config:       &aqua.Config{},
			enforceValue: false,
			defValue:     false,
			expected:     false,
		},
		{
			name: "require checksum is nil, use default",
			config: &aqua.Config{
				Checksum: &aqua.Checksum{},
			},
			enforceValue: false,
			defValue:     true,
			expected:     true,
		},
		{
			name: "require checksum is true",
			config: &aqua.Config{
				Checksum: &aqua.Checksum{
					RequireChecksum: boolPtr(true),
				},
			},
			enforceValue: false,
			defValue:     false,
			expected:     true,
		},
		{
			name: "require checksum is false",
			config: &aqua.Config{
				Checksum: &aqua.Checksum{
					RequireChecksum: boolPtr(false),
				},
			},
			enforceValue: false,
			defValue:     true,
			expected:     false,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.config.RequireChecksum(d.enforceValue, d.defValue)
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

func TestChecksum_GetEnabled(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		checksum *aqua.Checksum
		expected bool
	}{
		{
			name:     "checksum is nil",
			checksum: nil,
			expected: false,
		},
		{
			name:     "enabled is nil",
			checksum: &aqua.Checksum{},
			expected: false,
		},
		{
			name: "enabled is true",
			checksum: &aqua.Checksum{
				Enabled: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "enabled is false",
			checksum: &aqua.Checksum{
				Enabled: boolPtr(false),
			},
			expected: false,
		},
		{
			name: "full checksum configuration",
			checksum: &aqua.Checksum{
				Enabled:         boolPtr(true),
				RequireChecksum: boolPtr(false),
				SupportedEnvs:   registry.SupportedEnvs{"linux", "darwin"},
			},
			expected: true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.checksum.GetEnabled()
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}
