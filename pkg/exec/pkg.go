package exec

import (
	"context"
	"os/exec"
)

func (e *Executor) UnarchivePkg(ctx context.Context, pkgFilePath, dest string) (int, error) {
	return e.execAndOutputWhenFailure(e.command(exec.CommandContext(ctx, "pkgutil", "--expand-full", pkgFilePath, dest)))
}
