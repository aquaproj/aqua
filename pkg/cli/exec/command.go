// Package exec implements the aqua exec command for executing installed tools.
// The exec command is primarily used internally by aqua-proxy to execute tools
// installed by aqua, handling tool discovery, version resolution, and execution.
package exec

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// Args holds command-line arguments for the exec command.
type Args struct {
	*cliargs.GlobalArgs

	Command  string
	ExecArgs []string
}

// command holds the parameters and configuration for the exec command.
type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for executing tools.
// The returned command is used internally by aqua-proxy to execute
// installed tools with proper version resolution and argument handling.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
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
		Action: func(ctx context.Context, _ *cli.Command) error {
			return i.action(ctx, args)
		},
		ArgsUsage: `<executed command> [<arg> ...]`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "command",
				Destination: &args.Command,
			},
			&cli.StringArgs{
				Name:        "exec_args",
				Min:         0,
				Max:         -1,
				Destination: &args.ExecArgs,
			},
		},
	}
}

// action implements the main logic for the exec command.
// It parses the command arguments, initializes the exec controller,
// and executes the specified tool with the provided arguments.
func (i *command) action(ctx context.Context, args *Args) error {
	profiler, err := profile.Start(args.Trace, args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(args.GlobalArgs, i.r.Logger, param, i.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	ctrl, err := controller.InitializeExecCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize an ExecController: %w", err)
	}
	if args.Command == "" {
		return errors.New("command is required")
	}
	return ctrl.Exec(ctx, i.r.Logger.Logger, param, filepath.Base(args.Command), args.ExecArgs...) //nolint:wrapcheck
}
