package aqua

func (cfg *Config) ChecksumEnabled() bool {
	if cfg == nil {
		return false
	}
	if cfg.Checksum == nil {
		return true
	}
	if cfg.Checksum.Enabled == nil {
		return true
	}
	return *cfg.Checksum.Enabled
}

type Checksum struct {
	Enabled                  *bool              `json:"enabled,omitempty"`
	RequireChecksumInAdvance bool               `yaml:"require_checksum_in_advance" json:"require_checksum_in_advance,omitempty"`
	CreateJSON               bool               `yaml:"create_json" json:"create_json,omitempty"`
	RequireChecksum          bool               `yaml:"require_checksum" json:"require_checksum,omitempty"`
	SaveCalculatedChecksum   bool               `yaml:"save_calculated_checksum" json:"save_calculated_checksum,omitempty"`
	Excludes                 []*ChekcsumExclude `json:"excludes,omitempty"`
}

type ChekcsumExclude struct {
	Name     string   `json:"name,omitempty"`
	Registry string   `json:"registry,omitempty"`
	Version  string   `json:"version,omitempty"`
	Envs     []string `json:"envs,omitempty"`
}
