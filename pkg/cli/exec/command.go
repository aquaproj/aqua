// Package exec implements the aqua exec command for executing installed tools.
// The exec command is primarily used internally by aqua-proxy to execute tools
// installed by aqua, handling tool discovery, version resolution, and execution.
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

// command holds the parameters and configuration for the exec command.
type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for executing tools.
// The returned command is used internally by aqua-proxy to execute
// installed tools with proper version resolution and argument handling.
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

// action implements the main logic for the exec command.
// It parses the command arguments, initializes the exec controller,
// and executes the specified tool with the provided arguments.
func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, i.r.Logger, "exec", param, i.r.Version); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl, err := controller.InitializeExecCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize an ExecController: %w", err)
	}
	exeName, args, err := which.ParseExecArgs(cmd.Args().Slice())
	if err != nil {
		return fmt.Errorf("parse args: %w", err)
	}
	return ctrl.Exec(ctx, i.r.Logger.Logger, param, exeName, args...) //nolint:wrapcheck
}
