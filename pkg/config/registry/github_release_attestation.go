package registry

type GitHubReleaseAttestation struct {
	Enabled *bool `json:"enabled,omitempty"`
}

func (m *GitHubReleaseAttestation) GetEnabled() bool {
	if m == nil {
		return false
	}
	if m.Enabled != nil {
		return *m.Enabled
	}
	return true
}
