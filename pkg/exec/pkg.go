package exec

import (
	"context"
	"os/exec"
)

func (exe *Executor) UnarchivePkg(ctx context.Context, pkgFilePath, dest string) (int, error) {
	return exe.execAndOutputWhenFailure(exe.command(exec.CommandContext(ctx, "pkgutil", "--expand-full", pkgFilePath, dest)))
}
