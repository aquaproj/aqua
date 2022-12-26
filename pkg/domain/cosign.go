package domain

import (
	"context"

	"github.com/sirupsen/logrus"
)

type CosignInstaller interface {
	InstallCosign(ctx context.Context, logE *logrus.Entry, version string) error
}

type MockCosignInstaller struct {
	err error
}

func (mock *MockCosignInstaller) InstallCosign(ctx context.Context, logE *logrus.Entry, version string) error {
	return mock.err
}
