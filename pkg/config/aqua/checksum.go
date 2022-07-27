package aqua

func (cfg *Config) ChecksumEnabled() bool {
	if cfg == nil {
		return false
	}
	if cfg.Checksum == nil {
		return false
	}
	return cfg.Checksum.Enabled
}

type Checksum struct {
	Enabled bool `json:"enabled"`
}
