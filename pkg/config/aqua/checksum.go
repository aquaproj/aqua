package aqua

import "github.com/aquaproj/aqua/pkg/config/registry"

func (cfg *Config) ChecksumEnabled() bool {
	if cfg == nil {
		return false
	}
	return cfg.Checksum.GetEnabled()
}

func (cfg *Config) RequireChecksum(defValue bool) bool {
	if cfg == nil || cfg.Checksum == nil || cfg.Checksum.RequireChecksum == nil {
		return defValue
	}
	return *cfg.Checksum.RequireChecksum
}

type Checksum struct {
	Enabled         *bool                  `json:"enabled,omitempty"`
	RequireChecksum *bool                  `yaml:"require_checksum" json:"require_checksum,omitempty"`
	SupportedEnvs   registry.SupportedEnvs `yaml:"supported_envs" json:"supported_envs,omitempty"`
}
