package registry

type Checksum struct {
	Type       string           `json:"type"`
	Asset      string           `json:"asset,omitempty"`
	URL        string           `json:"url,omitempty"`
	FileFormat string           `yaml:"file_format" json:"file_format"`
	Algorithm  string           `json:"algorithm,omitempty"`
	Pattern    *ChecksumPattern `json:"pattern,omitempty"`
}

type ChecksumPattern struct {
	Checksum string `json:"checksum"`
	File     string `json:"file"`
}
