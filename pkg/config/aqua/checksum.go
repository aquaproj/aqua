package aqua

import "github.com/aquaproj/aqua/v2/pkg/config/registry"

func (c *Config) ChecksumEnabled(enforceValue, defValue bool) bool {
	if enforceValue {
		return true
	}
	if c == nil || c.Checksum == nil || c.Checksum.Enabled == nil {
		return defValue
	}
	return c.Checksum.GetEnabled()
}

func (c *Config) RequireChecksum(enforceValue, defValue bool) bool {
	if enforceValue {
		return true
	}
	if c == nil || c.Checksum == nil || c.Checksum.RequireChecksum == nil {
		return defValue
	}
	return *c.Checksum.RequireChecksum
}

type Checksum struct {
	Enabled         *bool                  `json:"enabled,omitempty"`
	RequireChecksum *bool                  `yaml:"require_checksum" json:"require_checksum,omitempty"`
	SupportedEnvs   registry.SupportedEnvs `yaml:"supported_envs" json:"supported_envs,omitempty"`
}
