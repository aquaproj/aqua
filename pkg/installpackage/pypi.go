package installpackage

import (
	"context"
	"fmt"
)

type PypiInstaller interface {
	Install(ctx context.Context, pkgName, target string) error
}

type MockPypiInstaller struct {
	Err error
}

func (mock *MockPypiInstaller) Install(ctx context.Context, pkgName, target string) error {
	return mock.Err
}

type PypiInstallerImpl struct {
	exec Executor
}

func NewPypiInstallerImpl(exec Executor) *PypiInstallerImpl {
	return &PypiInstallerImpl{
		exec: exec,
	}
}

func (inst *PypiInstallerImpl) Install(ctx context.Context, pkgName, target string) error {
	if _, err := inst.exec.Exec(ctx, "pip", "install", "--target", target, pkgName); err != nil {
		return fmt.Errorf("pip install: %w", err)
	}
	return nil
}
