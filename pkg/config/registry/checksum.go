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

func (chk *Checksum) GetCosign() *Cosign {
	if chk == nil {
		return nil
	}
	return chk.Cosign
}
