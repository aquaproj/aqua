package registry

import "fmt"

type SLSAProvenance struct {
	Enabled   *bool   `json:"enabled,omitempty"`
	Type      string  `json:"type,omitempty"  jsonschema:"enum=github_release,enum=http"`
	RepoOwner string  `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	RepoName  string  `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	Asset     *string `json:"asset,omitempty" yaml:",omitempty"`
	URL       *string `json:"url,omitempty" yaml:",omitempty"`
	SourceURI *string `json:"source_uri,omitempty" yaml:"source_uri"`
}

func (sp *SLSAProvenance) ToDownloadedFile() *DownloadedFile {
	return &DownloadedFile{
		Type:      sp.Type,
		RepoOwner: sp.RepoOwner,
		RepoName:  sp.RepoName,
		Asset:     sp.Asset,
		URL:       sp.URL,
	}
}

func (sp *SLSAProvenance) GetSourceURI() string {
	if sp.SourceURI != nil {
		return *sp.SourceURI
	}
	return fmt.Sprintf("github.com/%s/%s", sp.RepoOwner, sp.RepoName)
}

func (sp *SLSAProvenance) GetEnabled() bool {
	if sp == nil {
		return false
	}
	if sp.Enabled != nil {
		return *sp.Enabled
	}
	return sp.Type != ""
}

func (sp *SLSAProvenance) GetDownloadedFile() *DownloadedFile {
	return &DownloadedFile{
		Type:      sp.Type,
		RepoOwner: sp.RepoOwner,
		RepoName:  sp.RepoName,
		Asset:     sp.Asset,
		URL:       sp.URL,
	}
}
