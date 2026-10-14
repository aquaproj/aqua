package installpackage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
)

type GoBuildInstaller interface {
	Install(ctx context.Context, exePath, exeDir, src string, buildTags []string) error
}

type GoBuildInstallerImpl struct {
	exec Executor
}

func NewGoBuildInstallerImpl(exec Executor) *GoBuildInstallerImpl {
	return &GoBuildInstallerImpl{
		exec: exec,
	}
}

func (is *GoBuildInstallerImpl) Install(ctx context.Context, exePath, exeDir, src string, buildTags []string) error {
	args := []string{"build"}
	if len(buildTags) > 0 {
		args = append(args, "-tags", strings.Join(buildTags, ","))
	}
	args = append(args, "-o", exePath, src)
	cmd := osexec.Command(ctx, "go", args...)
	cmd.Dir = exeDir
	if _, err := is.exec.ExecStderr(cmd); err != nil {
		return fmt.Errorf("build a go package: %w", err)
	}
	return nil
}
