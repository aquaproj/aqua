package unarchive

import (
	"context"

	"github.com/sirupsen/logrus"
)

type MockUnarchiver struct {
	Err error
}

func (u *MockUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, dest string) error {
	return u.Err
}
