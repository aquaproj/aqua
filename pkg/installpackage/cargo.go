package installpackage

import (
	"context"
	"fmt"
)

type CargoPackageInstaller interface {
	Install(ctx context.Context, crate, version, root string) error
}

type MockCargoPackageInstaller struct {
	Err error
}

func (mock *MockCargoPackageInstaller) Install(ctx context.Context, crate, version, root string) error {
	return mock.Err
}

type CargoPackageInstallerImpl struct {
	exec Executor
}

func NewCargoPackageInstallerImpl(exec Executor) *CargoPackageInstallerImpl {
	return &CargoPackageInstallerImpl{
		exec: exec,
	}
}

func (inst *CargoPackageInstallerImpl) Install(ctx context.Context, crate, version, root string) error {
	_, err := inst.exec.Exec(ctx, "cargo", "install", "--version", version, "--root", root, crate)
	if err != nil {
		return fmt.Errorf("install a crate: %w", err)
	}
	return nil
}
