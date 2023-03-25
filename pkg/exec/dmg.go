package exec

import (
	"context"
	"os/exec"
)

func (exe *Executor) HdiutilDetach(ctx context.Context, mountPath string) (int, error) {
	cmd := exe.command(exec.Command("hdiutil", "detach", mountPath, "-quiet"))
	return exe.exec(ctx, cmd)
}

func (exe *Executor) HdiutilAttach(ctx context.Context, dmgPath, mountPoint string) (int, error) {
	cmd := exe.command(exec.Command("hdiutil", "attach", dmgPath, "-mountpoint", mountPoint, "-quiet"))
	return exe.exec(ctx, cmd)
}
