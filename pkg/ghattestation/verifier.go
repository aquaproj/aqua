package ghattestation

import (
	"context"
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

func (v *Verifier) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	return v.exe.Verify(ctx, logger, param) //nolint:wrapcheck
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
