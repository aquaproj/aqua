package ghattestation

import (
	"context"
	"log/slog"
)

func (v *Verifier) VerifyRelease(ctx context.Context, logger *slog.Logger, param *ParamVerifyRelease) error {
	return v.exe.VerifyRelease(ctx, logger, param) //nolint:wrapcheck
}
