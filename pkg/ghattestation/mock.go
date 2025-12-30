package ghattestation

import (
	"context"
	"log/slog"
)

type MockVerifier struct {
	err error
}

func (m *MockVerifier) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	return m.err
}

func (m *MockVerifier) VerifyRelease(ctx context.Context, logger *slog.Logger, param *ParamVerifyRelease) error {
	return m.err
}

type MockExecutor struct {
	Err error
}

func (m *MockExecutor) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify, signature string) error {
	return m.Err
}
