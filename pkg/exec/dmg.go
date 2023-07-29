package exec

import (
	"context"
	"os/exec"
)

func (exe *Executor) HdiutilDetach(ctx context.Context, mountPath string) (int, error) {
	cmd := exe.command(exec.CommandContext(ctx, "hdiutil", "detach", mountPath))
	return exe.execAndOutputWhenFailure(cmd)
}

func (exe *Executor) HdiutilAttach(ctx context.Context, dmgPath, mountPoint string) (int, error) {
	cmd := exe.command(exec.CommandContext(ctx, "hdiutil", "attach", dmgPath, "-mountpoint", mountPoint))
	return exe.execAndOutputWhenFailure(cmd)
}
