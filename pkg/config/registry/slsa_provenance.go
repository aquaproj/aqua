package registry

import "fmt"

// SLSAProvenance defines configuration for SLSA (Supply-chain Levels for Software Artifacts) provenance verification.
// SLSA is a framework for ensuring software supply chain security through build provenance.
type SLSAProvenance struct {
	// Enabled controls whether SLSA provenance verification is active.
	Enabled *bool `json:"enabled,omitempty"`
	// Type specifies where to download the provenance file from.
	Type string `json:"type,omitempty"  jsonschema:"enum=github_release,enum=http"`
	// RepoOwner is the GitHub repository owner (for github_release type).
	RepoOwner string `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	// RepoName is the GitHub repository name (for github_release type).
	RepoName string `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	// Asset is the name of the provenance file asset (for github_release type).
	Asset *string `json:"asset,omitempty" yaml:",omitempty"`
	// URL is the direct URL to the provenance file (for http type).
	URL *string `json:"url,omitempty" yaml:",omitempty"`
	// SourceURI is the expected source repository URI for verification.
	SourceURI *string `json:"source_uri,omitempty" yaml:"source_uri,omitempty"`
	// SourceTag is the expected source tag for verification.
	SourceTag string `json:"source_tag,omitempty" yaml:"source_tag,omitempty"`
}

// ToDownloadedFile converts the SLSAProvenance configuration to a DownloadedFile.
// This is used for downloading the provenance file.
func (sp *SLSAProvenance) ToDownloadedFile() *DownloadedFile {
	return &DownloadedFile{
		Type:      sp.Type,
		RepoOwner: sp.RepoOwner,
		RepoName:  sp.RepoName,
		Asset:     sp.Asset,
		URL:       sp.URL,
	}
}

// GetSourceURI returns the source URI for provenance verification.
// If SourceURI is not set, it derives it from RepoOwner and RepoName.
func (sp *SLSAProvenance) GetSourceURI() string {
	if sp.SourceURI != nil {
		return *sp.SourceURI
	}
	return fmt.Sprintf("github.com/%s/%s", sp.RepoOwner, sp.RepoName)
}

// GetEnabled returns whether SLSA provenance verification is enabled.
// If Enabled is nil, it's considered enabled if Type is configured.
func (sp *SLSAProvenance) GetEnabled() bool {
	if sp == nil {
		return false
	}
	if sp.Enabled != nil {
		return *sp.Enabled
	}
	return sp.Type != ""
}

// GetDownloadedFile returns a DownloadedFile for the provenance file.
// This is an alias for ToDownloadedFile for consistency with other verification types.
func (sp *SLSAProvenance) GetDownloadedFile() *DownloadedFile {
	return &DownloadedFile{
		Type:      sp.Type,
		RepoOwner: sp.RepoOwner,
		RepoName:  sp.RepoName,
		Asset:     sp.Asset,
		URL:       sp.URL,
	}
}
