package registry

type Minisign struct {
	Enabled   *bool   `json:"enabled,omitempty"`
	Type      string  `json:"type,omitempty"  jsonschema:"enum=github_release,enum=http"`
	RepoOwner string  `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	RepoName  string  `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	Asset     *string `json:"asset,omitempty" yaml:",omitempty"`
	URL       *string `json:"url,omitempty" yaml:",omitempty"`
	PublicKey string  `json:"public_key,omitempty" yaml:"public_key,omitempty"`
}

func (m *Minisign) ToDownloadedFile() *DownloadedFile {
	return &DownloadedFile{
		Type:      m.Type,
		RepoOwner: m.RepoOwner,
		RepoName:  m.RepoName,
		Asset:     m.Asset,
		URL:       m.URL,
	}
}

func (m *Minisign) GetEnabled() bool {
	if m == nil {
		return false
	}

	if m.Enabled != nil {
		return *m.Enabled
	}

	return true
}

func (m *Minisign) GetDownloadedFile() *DownloadedFile {
	return &DownloadedFile{
		Type:      m.Type,
		RepoOwner: m.RepoOwner,
		RepoName:  m.RepoName,
		Asset:     m.Asset,
		URL:       m.URL,
	}
}
