package registry

type Checksum struct {
	Type       string           `json:"type"`
	Path       string           `json:"path"`
	FileFormat string           `yaml:"file_format" json:"file_format"`
	Pattern    *ChecksumPattern `json:"pattern"`
}

type ChecksumPattern struct {
	Checksum string `json:"checksum"`
	File     string `json:"file"`
}
