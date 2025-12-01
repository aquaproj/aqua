package aqua

import "github.com/aquaproj/aqua/v2/pkg/config/registry"

// ChecksumEnabled determines if checksum validation is enabled.
// It considers enforcement flags and configuration settings to make the final decision.
func (c *Config) ChecksumEnabled(enforceValue, defValue bool) bool {
	if enforceValue {
		return true
	}
	if c == nil || c.Checksum == nil || c.Checksum.Enabled == nil {
		return defValue
	}
	return c.Checksum.GetEnabled()
}

// RequireChecksum determines if checksum validation is required.
// It considers enforcement flags and configuration to determine if checksums are mandatory.
func (c *Config) RequireChecksum(enforceValue, defValue bool) bool {
	if enforceValue {
		return true
	}
	if c == nil || c.Checksum == nil || c.Checksum.RequireChecksum == nil {
		return defValue
	}
	return *c.Checksum.RequireChecksum
}

// Checksum contains global checksum validation configuration.
// It controls whether checksums are enabled, required, and on which platforms.
type Checksum struct {
	Enabled         *bool                  `json:"enabled,omitempty"`                                  // Whether checksum validation is enabled
	RequireChecksum *bool                  `json:"require_checksum,omitempty" yaml:"require_checksum"` // Whether checksums are required for all packages
	SupportedEnvs   registry.SupportedEnvs `json:"supported_envs,omitempty"   yaml:"supported_envs"`   // Platforms where checksum validation applies
}
