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

func (m *MockGoBuildInstaller) Install(ctx context.Context, exePath, exeDir, src string) error {
	return m.Err
}

type GoBuildInstallerImpl struct {
	exec Executor
}

func NewGoBuildInstallerImpl(exec Executor) *GoBuildInstallerImpl {
	return &GoBuildInstallerImpl{
		exec: exec,
	}
}

func (is *GoBuildInstallerImpl) Install(ctx context.Context, exePath, exeDir, src string) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", exePath, src)
	cmd.Dir = exeDir
	if _, err := is.exec.ExecCommand(cmd); err != nil {
		return fmt.Errorf("build a go package: %w", err)
	}
	return nil
}
