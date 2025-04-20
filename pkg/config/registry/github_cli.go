package registry

type GitHubArtifactAttestations struct {
	Enabled       *bool  `json:"enabled,omitempty"`
	PredicateType string `json:"predicate_type,omitempty" yaml:"predicate_type,omitempty"`
	// https://github.com/aquaproj/aqua/issues/3581
	SignerWorkflow2 string `yaml:"signer_workflow,omitempty" json:"signer_workflow,omitempty"`
	// Deprecated: We'll remove signer-workflow at aqua v3.
	SignerWorkflow3 string `yaml:"signer-workflow,omitempty" json:"signer-workflow,omitempty"`
}

func (m *GitHubArtifactAttestations) SignerWorkflow() string {
	if m == nil {
		return ""
	}
	if m.SignerWorkflow2 != "" {
		return m.SignerWorkflow2
	}
	return m.SignerWorkflow3
}

func (m *GitHubArtifactAttestations) GetEnabled() bool {
	if m == nil {
		return false
	}
	if m.Enabled != nil {
		return *m.Enabled
	}
	return true
}
