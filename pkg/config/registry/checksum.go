package registry

type Checksum struct {
	Type         string           `json:"type,omitempty" jsonschema:"enum=github_release,enum=http"`
	Asset        string           `json:"asset,omitempty"`
	URL          string           `json:"url,omitempty"`
	FileFormat   string           `yaml:"file_format,omitempty" json:"file_format,omitempty"`
	Algorithm    string           `json:"algorithm,omitempty" jsonschema:"enum=md5,enum=sha1,enum=sha256,enum=sha512"`
	Pattern      *ChecksumPattern `json:"pattern,omitempty"`
	Enabled      *bool            `json:"enabled,omitempty"`
	Replacements Replacements     `json:"replacements,omitempty"`
	Cosign       *Cosign          `json:"cosign,omitempty"`
}

type ChecksumPattern struct {
	Checksum string `json:"checksum"`
	File     string `json:"file,omitempty"`
}

func (c *Checksum) GetReplacements() Replacements {
	if c == nil {
		return nil
	}
	return c.Replacements
}

func (c *Checksum) GetEnabled() bool {
	if c == nil {
		return false
	}
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

func (c *Checksum) GetAlgorithm() string {
	if !c.GetEnabled() {
		return "sha512"
	}
	return c.Algorithm
}

func (c *Checksum) GetCosign() *Cosign {
	if c == nil {
		return nil
	}
	return c.Cosign
}
