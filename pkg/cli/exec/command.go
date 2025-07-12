package exec

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/cli/which"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

type command struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}

	return &cli.Command{
		Name:  "exec",
		Usage: "Execute tool",
		Description: `Basically you don't have to use this command, because this is used by aqua internally. aqua-proxy invokes this command.
When you execute the command installed by aqua, "aqua exec" is executed internally.

e.g.
$ aqua exec -- gh version
gh version 2.4.0 (2021-12-21)
https://github.com/cli/cli/releases/tag/v2.4.0`,
		Action:    i.action,
		ArgsUsage: `<executed command> [<arg> ...]`,
	}
}

func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	err := util.SetParam(cmd, i.r.LogE, "exec", param, i.r.LDFlags)
	if err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl, err := controller.InitializeExecCommandController(ctx, i.r.LogE, param, http.DefaultClient, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize an ExecController: %w", err)
	}

	exeName, args, err := which.ParseExecArgs(cmd.Args().Slice())
	if err != nil {
		return fmt.Errorf("parse args: %w", err)
	}

	return ctrl.Exec(ctx, i.r.LogE, param, exeName, args...) //nolint:wrapcheck
}
