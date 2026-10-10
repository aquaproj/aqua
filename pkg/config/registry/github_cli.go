package registry

// GitHubArtifactAttestations defines configuration for GitHub artifact attestation verification.
// This uses GitHub's built-in attestation system for verifying build provenance and integrity.
type GitHubArtifactAttestations struct {
	// Enabled controls whether GitHub artifact attestation verification is active.
	Enabled *bool `yaml:",omitempty" json:"enabled,omitempty"`
	// PredicateType specifies the type of predicate to verify.
	PredicateType string `yaml:"predicate_type,omitempty" json:"predicate_type,omitempty"`
	// SignerWorkflow2 specifies the expected GitHub Actions workflow for signing.
	// It is a literal path such as owner/repo/.github/workflows/release.yml, not a
	// regular expression: gh >= v2.97.0 passes the value through regexp.QuoteMeta,
	// so any regex metacharacter in it is matched literally. Registries written for
	// older gh escaped the dots; ghattestation.unescapeSignerWorkflow undoes that,
	// so those values keep working, but new entries should be plain paths.
	// See https://github.com/aquaproj/aqua/issues/3581
	SignerWorkflow2 string `yaml:"signer_workflow,omitempty" json:"signer_workflow,omitempty"`
	// SignerWorkflow3 is the deprecated field name for signer workflow.
	//
	// Deprecated: Use SignerWorkflow2 instead. This will be removed in aqua v3.
	SignerWorkflow3 string `yaml:"signer-workflow,omitempty" json:"signer-workflow,omitempty" jsonschema:"description=Deprecated: use signer_workflow instead"`
	// SignerRepo is the repository whose workflow signed, passed as --signer-repo. It is for
	// a workflow in another repository than the artifact's -- a reusable one -- since
	// signer_workflow already names the repository otherwise.
	SignerRepo string `yaml:"signer_repo,omitempty" json:"signer_repo,omitempty"`
	// SourceRef is the ref the artifact was built from, such as refs/tags/v1.2.3, passed as
	// --source-ref.
	SourceRef string `yaml:"source_ref,omitempty" json:"source_ref,omitempty"`
	// SourceDigest is the commit the artifact was built from, passed as --source-digest.
	SourceDigest string `yaml:"source_digest,omitempty" json:"source_digest,omitempty"`
	// CertOIDCIssuer is the OIDC issuer the signing certificate must name, passed as
	// --cert-oidc-issuer.
	CertOIDCIssuer string `yaml:"cert_oidc_issuer,omitempty" json:"cert_oidc_issuer,omitempty"`
	// DenySelfHostedRunners refuses an attestation made on a self-hosted runner, passed as
	// --deny-self-hosted-runners.
	DenySelfHostedRunners bool `yaml:"deny_self_hosted_runners,omitempty" json:"deny_self_hosted_runners,omitempty"`
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
