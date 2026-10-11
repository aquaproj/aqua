package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
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

// The certificate fields are rendered like Opts, after them, so a definition can name every
// version's identity as a template.
func TestCosign_RenderOpts_certificate(t *testing.T) {
	t.Parallel()
	c := &registry.Cosign{
		Opts:                  []string{"--insecure-ignore-sct"},
		CertificateIdentity:   "https://github.com/foundry-rs/foundry/.github/workflows/release.yml@refs/tags/{{.Version}}",
		CertificateOIDCIssuer: "https://token.actions.githubusercontent.com",
	}
	got, err := c.RenderOpts(&runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}, &template.Artifact{Version: "v1.8.0"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"--insecure-ignore-sct",
		"--certificate-identity", "https://github.com/foundry-rs/foundry/.github/workflows/release.yml@refs/tags/v1.8.0",
		"--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}
