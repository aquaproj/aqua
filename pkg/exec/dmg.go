package exec

import (
	"context"
	"os/exec"
)

func (exe *Executor) HdiutilDetach(ctx context.Context, mountPath string) (int, error) {
	cmd := exe.command(exec.Command("hdiutil", "detach", mountPath))
	return exe.execAndOutputWhenFailure(ctx, cmd)
}

func (exe *Executor) HdiutilAttach(ctx context.Context, dmgPath, mountPoint string) (int, error) {
	cmd := exe.command(exec.Command("hdiutil", "attach", dmgPath, "-mountpoint", mountPoint))
	return exe.execAndOutputWhenFailure(ctx, cmd)
}
