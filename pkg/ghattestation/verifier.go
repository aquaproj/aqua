package ghattestation

import (
	"context"
	"errors"
	"log/slog"
)

type Verifier struct {
	exe Executor
}

func New(exe Executor) *Verifier {
	return &Verifier{
		exe: exe,
	}
}

type ParamVerify struct {
	ArtifactPath          string
	Repository            string
	SignerWorkflow        string
	PredicateType         string
	SignerRepo            string
	SourceRef             string
	SourceDigest          string
	CertOIDCIssuer        string
	DenySelfHostedRunners bool
}

var errSignerRepoAndWorkflow = errors.New("signer_repo and signer_workflow can't be set together, because gh attestation verify accepts only one of them")

func (v *Verifier) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	// gh refuses the combination as a flag error, which would otherwise be retried and
	// reported as a verification failure.
	if err := param.validate(); err != nil {
		return err
	}
	return v.exe.Verify(ctx, logger, param) //nolint:wrapcheck
}

// validate rejects what gh attestation verify would refuse as flags rather than as an
// attestation.
func (p *ParamVerify) validate() error {
	if p.SignerRepo != "" && p.SignerWorkflow != "" {
		return errSignerRepoAndWorkflow
	}
	return nil
}

// certificateArgs returns the flags pinning what the signing certificate says, in a fixed
// order. Each is passed only when it is set.
func (p *ParamVerify) certificateArgs() []string {
	fields := []struct{ flag, value string }{
		{"--signer-repo", p.SignerRepo},
		{"--source-ref", p.SourceRef},
		{"--source-digest", p.SourceDigest},
		{"--cert-oidc-issuer", p.CertOIDCIssuer},
	}
	args := []string{}
	for _, f := range fields {
		if f.value != "" {
			args = append(args, f.flag, f.value)
		}
	}
	if p.DenySelfHostedRunners {
		args = append(args, "--deny-self-hosted-runners")
	}
	return args
}
