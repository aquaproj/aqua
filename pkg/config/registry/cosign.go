package registry

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

// Cosign defines configuration for verifying packages using Cosign signature verification.
// Cosign is a tool for signing and verifying container images and other artifacts.
type Cosign struct {
	// Enabled controls whether Cosign verification is active.
	Enabled *bool `yaml:",omitempty" json:"enabled,omitempty"`
	// Opts contains additional command-line options to pass to cosign verify.
	Opts []string `yaml:",omitempty" json:"opts,omitempty"`
	// Signature specifies where to download the signature file.
	Signature *DownloadedFile `yaml:",omitempty" json:"signature,omitempty"`
	// Certificate specifies where to download the certificate file.
	Certificate *DownloadedFile `yaml:",omitempty" json:"certificate,omitempty"`
	// Key specifies where to download the public key file.
	Key *DownloadedFile `yaml:",omitempty" json:"key,omitempty"`
	// Bundle specifies where to download the signature bundle.
	Bundle *DownloadedFile `yaml:",omitempty" json:"bundle,omitempty"`
}

// DownloadedFile represents a file that can be downloaded from various sources.
// This is used for signature files, certificates, and other verification artifacts.
type DownloadedFile struct {
	// Type specifies the source type for downloading the file.
	Type string `json:"type" jsonschema:"enum=github_release,enum=http,enum=forgejo_release,enum=gitea_release,enum=gitlab_release"`
	// RepoOwner is the repository owner (for github_release, forgejo_release, gitea_release and gitlab_release types).
	RepoOwner string `yaml:"repo_owner,omitempty" json:"repo_owner,omitempty"`
	// RepoName is the repository name (for github_release, forgejo_release, gitea_release and gitlab_release types).
	RepoName string `yaml:"repo_name,omitempty" json:"repo_name,omitempty"`
	// Asset is the name of the asset to download (for github_release,
	// forgejo_release, gitea_release and gitlab_release types). A file on a forge instance is on the
	// instance the package itself is on: a signature is published beside what it
	// signs.
	Asset *string `yaml:",omitempty" json:"asset,omitempty"`
	// URL is the direct URL to download the file (for http type).
	URL *string `yaml:",omitempty" json:"url,omitempty"`
}

// GetEnabled returns whether Cosign verification is enabled.
// If Enabled is nil, it's considered enabled if any verification files or options are configured.
func (c *Cosign) GetEnabled() bool {
	if c == nil {
		return false
	}
	if c.Enabled != nil {
		return *c.Enabled
	}
	return len(c.Opts) != 0 || c.Signature != nil || c.Certificate != nil || c.Key != nil || c.Bundle != nil
}

// RenderOpts renders the Cosign command-line options with template substitution.
// It replaces template variables in the options with runtime and artifact values.
func (c *Cosign) RenderOpts(rt *runtime.Runtime, art *template.Artifact) ([]string, error) {
	opts := make([]string, len(c.Opts))
	for i, opt := range c.Opts {
		s, err := template.Render(opt, art, rt)
		if err != nil {
			return nil, fmt.Errorf("render a cosign option: %w", err)
		}
		opts[i] = s
	}

	return opts, nil
}
