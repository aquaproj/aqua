package exec

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/cli/which"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
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

func (i *command) action(c *cli.Context) error {
	profiler, err := profile.Start(c)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	logE := i.r.LogE

	param := &config.Param{}
	if err := util.SetParam(c, logE, "exec", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl, err := controller.InitializeExecCommandController(c.Context, param, http.DefaultClient, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize a ExecController: %w", err)
	}
	exeName, args, err := which.ParseExecArgs(c.Args().Slice())
	if err != nil {
		return fmt.Errorf("parse args: %w", err)
	}
	defer ctrl.CloseVacuum(logE)
	return ctrl.Exec(c.Context, logE, param, exeName, args...) //nolint:wrapcheck
}
