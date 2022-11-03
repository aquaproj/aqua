package registry

type Checksum struct {
	Type         string           `json:"type,omitempty"`
	Asset        string           `json:"asset,omitempty"`
	URL          string           `json:"url,omitempty"`
	FileFormat   string           `yaml:"file_format" json:"file_format,omitempty"`
	Algorithm    string           `json:"algorithm,omitempty"`
	Pattern      *ChecksumPattern `json:"pattern,omitempty"`
	Enabled      *bool            `json:"enabled,omitempty"`
	Replacements Replacements     `json:"replacements,omitempty"`
}

type ChecksumPattern struct {
	Checksum string `json:"checksum"`
	File     string `json:"file,omitempty"`
}

func (chk *Checksum) GetReplacements() Replacements {
	if chk == nil {
		return nil
	}
	return chk.Replacements
}

func (chk *Checksum) GetEnabled() bool {
	if chk == nil {
		return false
	}
	if chk.Enabled == nil {
		return true
	}
	return *chk.Enabled
}

func (chk *Checksum) GetAlgorithm() string {
	if !chk.GetEnabled() {
		return "sha512"
	}
	return chk.Algorithm
}
