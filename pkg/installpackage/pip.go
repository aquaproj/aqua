package installpackage

import (
	"context"
	"fmt"
)

type PipInstaller interface {
	Install(ctx context.Context, pkgName, target string) error
}

type MockPipInstaller struct {
	Err error
}

func (mock *MockPipInstaller) Install(ctx context.Context, pkgName, target string) error {
	return mock.Err
}

type PipInstallerImpl struct {
	exec Executor
}

func NewPipInstallerImpl(exec Executor) *PipInstallerImpl {
	return &PipInstallerImpl{
		exec: exec,
	}
}

func (inst *PipInstallerImpl) Install(ctx context.Context, pkgName, target string) error {
	if _, err := inst.exec.Exec(ctx, "pip", "install", "--target", target, pkgName); err != nil {
		return fmt.Errorf("pip install: %w", err)
	}
	return nil
}
