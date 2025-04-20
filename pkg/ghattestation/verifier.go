package ghattestation

import (
	"context"

	"github.com/sirupsen/logrus"
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

func (v *Verifier) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error {
	return v.exe.Verify(ctx, logE, param) //nolint:wrapcheck
}
