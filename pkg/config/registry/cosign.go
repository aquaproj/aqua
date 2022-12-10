package registry

type Cosign struct {
	Enabled            *bool           `json:"enabled"`
	CosignExperimental bool            `yaml:"cosign_experimental" json:"cosign_experimental,omitempty"`
	Opts               []string        `json:"opts,omitempty"`
	Signature          *DownloadedFile `json:"signature,omitempty"`
	Certificate        *DownloadedFile `json:"certificate,omitempty"`
	Key                *DownloadedFile `json:"key,omitempty"`
}

type DownloadedFile struct {
	Type      string  `validate:"required" json:"type" jsonschema:"enum=github_release,enum=http"`
	RepoOwner string  `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	RepoName  string  `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	Asset     *string `json:"asset,omitempty" yaml:",omitempty"`
	URL       *string `json:"url,omitempty" yaml:",omitempty"`
}

func (cos *Cosign) GetEnabled() bool {
	if cos == nil {
		return false
	}
	if cos.Enabled != nil {
		return *cos.Enabled
	}
	return len(cos.Opts) != 0 || cos.Signature != nil || cos.Certificate != nil || cos.Key != nil || cos.CosignExperimental
}
