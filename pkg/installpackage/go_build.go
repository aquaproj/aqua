package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
)

type GoBuildInstaller interface {
	Install(ctx context.Context, exePath, exeDir, src string) error
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
	cmd := osexec.Command(ctx, "go", "build", "-o", exePath, src)

	cmd.Dir = exeDir
	if _, err := is.exec.ExecStderr(cmd); err != nil {
		return fmt.Errorf("build a go package: %w", err)
	}

	return nil
}
