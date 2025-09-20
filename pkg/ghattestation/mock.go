package ghattestation

import (
	"context"

	"github.com/sirupsen/logrus"
)

type MockVerifier struct {
	err error
}

func (m *MockVerifier) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error {
	return m.err
}

func (m *MockVerifier) VerifyRelease(ctx context.Context, logE *logrus.Entry, param *ParamVerifyRelease) error {
	return m.err
}

type MockExecutor struct {
	Err error
}

func (m *MockExecutor) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, signature string) error {
	return m.Err
}
