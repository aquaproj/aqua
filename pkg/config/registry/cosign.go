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
	// CertificateIdentity is the identity the signing certificate must carry, passed as
	// --certificate-identity.
	CertificateIdentity string `yaml:"certificate_identity,omitempty" json:"certificate_identity,omitempty"`
	// CertificateIdentityRegexp is a regular expression the certificate's identity must
	// match, passed as --certificate-identity-regexp.
	CertificateIdentityRegexp string `yaml:"certificate_identity_regexp,omitempty" json:"certificate_identity_regexp,omitempty"`
	// CertificateOIDCIssuer is the OIDC issuer the certificate must name, passed as
	// --certificate-oidc-issuer.
	CertificateOIDCIssuer string `yaml:"certificate_oidc_issuer,omitempty" json:"certificate_oidc_issuer,omitempty"`
	// CertificateGitHubWorkflowRepository is the repository whose workflow signed, passed as
	// --certificate-github-workflow-repository.
	CertificateGitHubWorkflowRepository string `yaml:"certificate_github_workflow_repository,omitempty" json:"certificate_github_workflow_repository,omitempty"`
	// CertificateGitHubWorkflowRef is the ref that workflow ran on, passed as
	// --certificate-github-workflow-ref.
	CertificateGitHubWorkflowRef string `yaml:"certificate_github_workflow_ref,omitempty" json:"certificate_github_workflow_ref,omitempty"`
}

// CertificateOpts returns the flags the certificate fields say, in a fixed order.
//
// They are fields rather than a part of Opts so that what is verified can be read without
// parsing a command line. Opts still carries whatever no field names.
func (c *Cosign) CertificateOpts() []string {
	if c == nil {
		return nil
	}
	fields := []struct{ flag, value string }{
		{"--certificate-identity", c.CertificateIdentity},
		{"--certificate-identity-regexp", c.CertificateIdentityRegexp},
		{"--certificate-oidc-issuer", c.CertificateOIDCIssuer},
		{"--certificate-github-workflow-repository", c.CertificateGitHubWorkflowRepository},
		{"--certificate-github-workflow-ref", c.CertificateGitHubWorkflowRef},
	}
	opts := []string{}
	for _, f := range fields {
		if f.value != "" {
			opts = append(opts, f.flag, f.value)
		}
	}
	return opts
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
