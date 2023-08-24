package exec

import (
	"context"
	"os/exec"
)

func (e *Executor) HdiutilDetach(ctx context.Context, mountPath string) (int, error) {
	cmd := e.command(exec.CommandContext(ctx, "hdiutil", "detach", mountPath))
	return e.execAndOutputWhenFailure(cmd)
}

func (e *Executor) HdiutilAttach(ctx context.Context, dmgPath, mountPoint string) (int, error) {
	cmd := e.command(exec.CommandContext(ctx, "hdiutil", "attach", dmgPath, "-mountpoint", mountPoint))
	return e.execAndOutputWhenFailure(cmd)
}
