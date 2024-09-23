package registry

type GitHubArtifactAttestations struct {
	Enabled        *bool  `json:"enabled,omitempty"`
	SignerWorkflow string `yaml:"signer-workflow,omitempty" json:"signer-workflow,omitempty"`
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
