package installpackage

import (
	"context"
	"fmt"
)

type PypiInstaller interface {
	Install(ctx context.Context, pkgName, version, target string) error
}

type MockPypiInstaller struct {
	Err error
}

func (mock *MockPypiInstaller) Install(ctx context.Context, pkgName, version, target string) error {
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

func (inst *PypiInstallerImpl) Install(ctx context.Context, pkgName, version, target string) error {
	// https://stackoverflow.com/questions/50821312/what-is-the-effect-of-using-python-m-pip-instead-of-just-pip
	if _, err := inst.exec.Exec(ctx, "python", "-m", "pip", "install", "--target", target, fmt.Sprintf("%s==%s", pkgName, version)); err != nil {
		return fmt.Errorf("pip install: %w", err)
	}
	return nil
}
