package installpackage

import (
	"context"
)

type MockGoBuildInstaller struct {
	Err error
}

func (m *MockGoBuildInstaller) Install(ctx context.Context, exePath, exeDir, src string) error {
	return m.Err
}
