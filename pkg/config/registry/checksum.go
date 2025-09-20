package registry

// Checksum defines configuration for verifying package integrity using checksums.
// It supports downloading checksum files from various sources and multiple hash algorithms.
type Checksum struct {
	// Type specifies where to download the checksum file from.
	Type string `json:"type,omitempty" jsonschema:"enum=github_release,enum=http"`
	// Asset is the name of the checksum file asset (for github_release type).
	Asset string `json:"asset,omitempty"`
	// URL is the direct URL to the checksum file (for http type).
	URL string `json:"url,omitempty"`
	// FileFormat specifies the format of the checksum file.
	FileFormat string `yaml:"file_format,omitempty" json:"file_format,omitempty"`
	// Algorithm specifies the hash algorithm used for checksums.
	Algorithm string `json:"algorithm,omitempty" jsonschema:"enum=md5,enum=sha1,enum=sha256,enum=sha512"`
	// Pattern defines how to extract checksums from the checksum file.
	Pattern *ChecksumPattern `json:"pattern,omitempty"`
	// Enabled controls whether checksum verification is active.
	Enabled *bool `json:"enabled,omitempty"`
	// Replacements provides template replacements for checksum URLs/assets.
	Replacements Replacements `json:"replacements,omitempty"`
	// Cosign configuration for signature verification of checksums.
	Cosign *Cosign `json:"cosign,omitempty"`
	// Minisign configuration for signature verification of checksums.
	Minisign *Minisign `json:"minisign,omitempty"`
	// GitHubArtifactAttestations configuration for GitHub artifact attestation verification.
	GitHubArtifactAttestations *GitHubArtifactAttestations `json:"github_artifact_attestations,omitempty" yaml:"github_artifact_attestations,omitempty"`
}

// ChecksumPattern defines regular expression patterns for extracting checksums from checksum files.
// This is used when checksum files contain multiple checksums in a specific format.
type ChecksumPattern struct {
	// Checksum is a regex pattern to extract the checksum value.
	Checksum string `json:"checksum"`
	// File is a regex pattern to extract the filename (optional).
	File string `json:"file,omitempty"`
}

// GetReplacements returns the template replacements for this checksum configuration.
// It returns nil if the checksum is nil.
func (c *Checksum) GetReplacements() Replacements {
	if c == nil {
		return nil
	}
	return c.Replacements
}

// GetEnabled returns whether checksum verification is enabled.
// If Enabled is nil, it defaults to true.
func (c *Checksum) GetEnabled() bool {
	if c == nil {
		return false
	}
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

// GetAlgorithm returns the hash algorithm to use for checksum verification.
// If checksum is disabled, it returns "sha256" as a default.
func (c *Checksum) GetAlgorithm() string {
	if !c.GetEnabled() {
		return "sha256"
	}
	return c.Algorithm
}

// GetCosign returns the Cosign configuration for signature verification.
// It returns nil if the checksum is nil.
func (c *Checksum) GetCosign() *Cosign {
	if c == nil {
		return nil
	}
	return c.Cosign
}

// GetMinisign returns the Minisign configuration for signature verification.
// It returns nil if the checksum is nil.
func (c *Checksum) GetMinisign() *Minisign {
	if c == nil {
		return nil
	}
	return c.Minisign
}

// GetGitHubArtifactAttestations returns the GitHub artifact attestation configuration.
// It returns nil if the checksum is nil.
func (c *Checksum) GetGitHubArtifactAttestations() *GitHubArtifactAttestations {
	if c == nil {
		return nil
	}
	return c.GitHubArtifactAttestations
}
