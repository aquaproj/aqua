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
	Enabled *bool `json:"enabled,omitempty"`
	// CreateJSON               bool               `yaml:"create_json" json:"-"`
	RequireChecksum *bool `yaml:"require_checksum" json:"require_checksum,omitempty"`
	// SaveCalculatedChecksum   bool               `yaml:"save_calculated_checksum" json:"-"`
	// Excludes                 []*ChekcsumExclude `json:"-"`
	SupportedEnvs registry.SupportedEnvs `yaml:"supported_envs" json:"supported_envs,omitempty"`
}

// type ChekcsumExclude struct {
// 	Name     string   `json:"name,omitempty"`
// 	Registry string   `json:"registry,omitempty"`
// 	Version  string   `json:"version,omitempty"`
// 	Envs     []string `json:"envs,omitempty"`
// }
