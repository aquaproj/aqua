package registry

// GitHubArtifactAttestations defines configuration for GitHub artifact attestation verification.
// This uses GitHub's built-in attestation system for verifying build provenance and integrity.
type GitHubArtifactAttestations struct {
	// Enabled controls whether GitHub artifact attestation verification is active.
	Enabled *bool `json:"enabled,omitempty"`
	// PredicateType specifies the type of predicate to verify.
	PredicateType string `json:"predicate_type,omitempty" yaml:"predicate_type,omitempty"`
	// SignerWorkflow2 specifies the expected GitHub Actions workflow for signing.
	// See https://github.com/aquaproj/aqua/issues/3581
	SignerWorkflow2 string `yaml:"signer_workflow,omitempty" json:"signer_workflow,omitempty"`
	// SignerWorkflow3 is the deprecated field name for signer workflow.
	// Deprecated: Use SignerWorkflow2 instead. This will be removed in aqua v3.
	SignerWorkflow3 string `yaml:"signer-workflow,omitempty" json:"signer-workflow,omitempty"`
}

// SignerWorkflow returns the configured signer workflow.
// It prefers SignerWorkflow2 over the deprecated SignerWorkflow3.
func (m *GitHubArtifactAttestations) SignerWorkflow() string {
	if m == nil {
		return ""
	}
	if m.SignerWorkflow2 != "" {
		return m.SignerWorkflow2
	}
	return m.SignerWorkflow3
}

// GetEnabled returns whether GitHub artifact attestation verification is enabled.
// If Enabled is nil, it defaults to true.
func (m *GitHubArtifactAttestations) GetEnabled() bool {
	if m == nil {
		return false
	}
	if m.Enabled != nil {
		return *m.Enabled
	}
	return true
}
