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
	ArtifactPath   string
	Repository     string
	SignerWorkflow string
	PredicateType  string
}

func (v *Verifier) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	return v.exe.Verify(ctx, logger, param) //nolint:wrapcheck
}
