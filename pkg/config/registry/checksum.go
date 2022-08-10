package registry

type Checksum struct {
	Type       string           `json:"type"`
	Asset      string           `json:"asset,omitempty"`
	URL        string           `json:"url,omitempty"`
	FileFormat string           `yaml:"file_format" json:"file_format"`
	Algorithm  string           `json:"algorithm,omitempty"`
	Pattern    *ChecksumPattern `json:"pattern,omitempty"`
	Disabled   bool             `json:"disabled,omitempty"`
}

type ChecksumPattern struct {
	Checksum string `json:"checksum"`
	File     string `json:"file"`
}

func (chk *Checksum) Enabled() bool {
	if chk == nil {
		return false
	}
	return !chk.Disabled
}

func (chk *Checksum) GetAlgorithm() string {
	if chk == nil {
		return "sha256"
	}
	return chk.Algorithm
}
