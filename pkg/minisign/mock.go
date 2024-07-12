package minisign

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

type MockExecutor struct {
	Err error
}

func (m *MockExecutor) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error {
	return m.Err
}
