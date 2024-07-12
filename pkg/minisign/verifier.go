package minisign

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Verifier struct {
	fs  afero.Fs
	exe Executor
}

func New(fs afero.Fs, exe Executor) *Verifier {
	return &Verifier{
		fs:  fs,
		exe: exe,
	}
}

type ParamVerify struct {
	ArtifactPath string
	PublicKey    string
}

func (v *Verifier) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error {
	return v.exe.Verify(ctx, logE, param) //nolint:wrapcheck
}
