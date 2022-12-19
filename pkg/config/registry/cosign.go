package registry

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
)

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

func (cos *Cosign) RenderOpts(rt *runtime.Runtime, art *template.Artifact) ([]string, error) {
	opts := make([]string, len(cos.Opts))
	for i, opt := range cos.Opts {
		s, err := template.Render(opt, art, rt)
		if err != nil {
			return nil, fmt.Errorf("render a cosign option: %w", err)
		}
		opts[i] = s
	}

	return opts, nil
}
