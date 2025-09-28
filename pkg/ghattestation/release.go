package ghattestation

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (v *Verifier) VerifyRelease(ctx context.Context, logE *logrus.Entry, param *ParamVerifyRelease) error {
	return v.exe.VerifyRelease(ctx, logE, param) //nolint:wrapcheck
}
