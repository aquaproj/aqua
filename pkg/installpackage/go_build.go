package installpackage

import (
	"context"
	"fmt"
	"os/exec"
)

type GoBuildInstaller interface {
	Install(ctx context.Context, exePath, exeDir, src string) error
}

type MockGoBuildInstaller struct {
	Err error
}

func (mock *MockGoBuildInstaller) Install(ctx context.Context, exePath, exeDir, src string) error {
	return mock.Err
}

type GoBuildInstallerImpl struct {
	exec Executor
}

func NewGoBuildInstallerImpl(exec Executor) *GoBuildInstallerImpl {
	return &GoBuildInstallerImpl{
		exec: exec,
	}
}

func (inst *GoBuildInstallerImpl) Install(ctx context.Context, exePath, exeDir, src string) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", exePath, src)
	cmd.Dir = exeDir
	if _, err := inst.exec.ExecCommand(cmd); err != nil {
		return fmt.Errorf("build a go package: %w", err)
	}
	return nil
}
