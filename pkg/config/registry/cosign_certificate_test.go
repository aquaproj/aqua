package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

// The certificate fields are flags in a fixed order, each only when it is set, so the
// command a verification runs is the same every time for the same entry.
func TestCosign_CertificateOpts(t *testing.T) {
	t.Parallel()
	c := &registry.Cosign{
		CertificateGitHubWorkflowRef:        "refs/tags/v1.8.0",
		CertificateIdentity:                 "https://github.com/foundry-rs/foundry/.github/workflows/release.yml@refs/tags/v1.8.0",
		CertificateOIDCIssuer:               "https://token.actions.githubusercontent.com",
		CertificateGitHubWorkflowRepository: "foundry-rs/foundry",
	}
	want := []string{
		"--certificate-identity", "https://github.com/foundry-rs/foundry/.github/workflows/release.yml@refs/tags/v1.8.0",
		"--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
		"--certificate-github-workflow-repository", "foundry-rs/foundry",
		"--certificate-github-workflow-ref", "refs/tags/v1.8.0",
	}
	if diff := cmp.Diff(want, c.CertificateOpts()); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	if got := (&registry.Cosign{}).CertificateOpts(); len(got) != 0 {
		t.Errorf("no field set gave %v", got)
	}
	if got := (*registry.Cosign)(nil).CertificateOpts(); got != nil {
		t.Errorf("nil gave %v", got)
	}
}
