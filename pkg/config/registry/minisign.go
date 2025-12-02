package registry

// Minisign defines configuration for verifying packages using Minisign signature verification.
// Minisign is a simple tool for signing files and verifying signatures.
type Minisign struct {
	// Enabled controls whether Minisign verification is active.
	Enabled *bool `json:"enabled,omitempty"`
	// Type specifies where to download the signature file from.
	Type string `json:"type,omitempty" jsonschema:"enum=github_release,enum=http"`
	// RepoOwner is the GitHub repository owner (for github_release type).
	RepoOwner string `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	// RepoName is the GitHub repository name (for github_release type).
	RepoName string `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	// Asset is the name of the signature file asset (for github_release type).
	Asset *string `yaml:",omitempty" json:"asset,omitempty"`
	// URL is the direct URL to the signature file (for http type).
	URL *string `yaml:",omitempty" json:"url,omitempty"`
	// PublicKey is the base64-encoded public key for verification.
	PublicKey string `yaml:"public_key,omitempty" json:"public_key,omitempty"`
}

// ToDownloadedFile converts the Minisign configuration to a DownloadedFile.
// This is used for downloading the signature file.
func (m *Minisign) ToDownloadedFile() *DownloadedFile {
	return &DownloadedFile{
		Type:      m.Type,
		RepoOwner: m.RepoOwner,
		RepoName:  m.RepoName,
		Asset:     m.Asset,
		URL:       m.URL,
	}
}

// GetEnabled returns whether Minisign verification is enabled.
// If Enabled is nil, it defaults to true.
func (m *Minisign) GetEnabled() bool {
	if m == nil {
		return false
	}
	if m.Enabled != nil {
		return *m.Enabled
	}
	return true
}
