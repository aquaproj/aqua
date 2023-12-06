package installpackage

import (
	"context"
)

type MockGoInstallInstaller struct {
	Err error
}

func (m *MockGoInstallInstaller) Install(ctx context.Context, path, gobin string) error {
	return m.Err
}
