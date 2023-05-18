package exec

import (
	"context"
	"os/exec"
)

func (exe *Executor) UnarchivePkg(ctx context.Context, pkgFilePath, dest string) (int, error) {
	cmd := exe.command(exec.Command("pkgutil", "--expand-full", pkgFilePath, dest))
	return exe.execAndOutputWhenFailure(ctx, cmd)
}
