package ghattestation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// What the certificate is pinned to becomes flags in a fixed order, each only when set.
func TestParamVerify_certificateArgs(t *testing.T) {
	t.Parallel()
	p := &ParamVerify{
		SourceDigest:          "0123456789abcdef0123456789abcdef01234567",
		SourceRef:             "refs/tags/v1.2.3",
		CertOIDCIssuer:        "https://token.actions.githubusercontent.com",
		SignerRepo:            "org/workflows",
		DenySelfHostedRunners: true,
	}
	want := []string{
		"--signer-repo", "org/workflows",
		"--source-ref", "refs/tags/v1.2.3",
		"--source-digest", "0123456789abcdef0123456789abcdef01234567",
		"--cert-oidc-issuer", "https://token.actions.githubusercontent.com",
		"--deny-self-hosted-runners",
	}
	if diff := cmp.Diff(want, p.certificateArgs()); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	if got := (&ParamVerify{}).certificateArgs(); len(got) != 0 {
		t.Errorf("no field set gave %v", got)
	}
}

// gh refuses signer_repo together with signer_workflow, so aqua refuses it before running gh.
func TestParamVerify_validate(t *testing.T) {
	t.Parallel()
	both := &ParamVerify{SignerRepo: "org/workflows", SignerWorkflow: "org/workflows/.github/workflows/release.yaml"}
	if err := both.validate(); err == nil {
		t.Error("signer_repo with signer_workflow was accepted")
	}
	for _, p := range []*ParamVerify{
		{SignerRepo: "org/workflows"},
		{SignerWorkflow: "org/workflows/.github/workflows/release.yaml"},
		{},
	} {
		if err := p.validate(); err != nil {
			t.Errorf("%+v: %v", p, err)
		}
	}
}
