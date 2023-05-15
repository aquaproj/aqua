package installpackage

import (
	"context"
	"fmt"
)

type GoInstallInstaller interface {
	Install(ctx context.Context, path, gobin string) error
}

type MockGoInstallInstaller struct {
	Err error
}

func (mock *MockGoInstallInstaller) Install(ctx context.Context, path, gobin string) error {
	return mock.Err
}

type GoInstallInstallerImpl struct {
	exec Executor
}

func NewGoInstallInstallerImpl(exec Executor) *GoInstallInstallerImpl {
	return &GoInstallInstallerImpl{
		exec: exec,
	}
}

func (inst *GoInstallInstallerImpl) Install(ctx context.Context, path, gobin string) error {
	_, err := inst.exec.ExecWithEnvs(ctx, "go", []string{"install", path}, []string{fmt.Sprintf("GOBIN=%s", gobin)})
	return fmt.Errorf("install a go package: %w", err)
}
