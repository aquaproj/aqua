package registry

type Checksum struct {
	Type       string           `json:"type"`
	Path       string           `json:"path,omitempty"`
	FileFormat string           `yaml:"file_format" json:"file_format"`
	Pattern    *ChecksumPattern `json:"pattern,omitempty"`
}

type ChecksumPattern struct {
	Checksum string `json:"checksum"`
	File     string `json:"file"`
}
