package unarchive

import (
	"context"
	"log/slog"
)

type MockUnarchiver struct {
	Err error
}

func (u *MockUnarchiver) Unarchive(_ context.Context, _ *slog.Logger, _ *File, _ string) error {
	return u.Err
}
