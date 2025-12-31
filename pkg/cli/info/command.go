// Package info implements the aqua info command for displaying system information.
// The info command shows various details about the aqua installation,
// configuration, and environment for debugging and troubleshooting purposes.
package info

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// Args holds command-line arguments for the info command.
type Args struct {
	*cliargs.GlobalArgs
}

// infoCommand holds the parameters and configuration for the info command.
type infoCommand struct {
	r    *util.Param
	args *Args
}

// New creates and returns a new CLI command for displaying information.
// The returned command shows system information about aqua installation,
// configuration paths, and environment details for troubleshooting.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &infoCommand{
		r:    r,
		args: args,
	}
	return &cli.Command{
		Name:  "info",
		Usage: "Show information",
		Description: `Show information.
e.g.
$ aqua info`,
		Action: i.action,
		Flags:  []cli.Flag{},
	}
}

// action implements the main logic for the info command.
// It initializes the info controller and displays comprehensive
// information about the aqua installation and configuration.
func (i *infoCommand) action(ctx context.Context, _ *cli.Command) error {
	profiler, err := profile.Start(i.args.Trace, i.args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(i.args.GlobalArgs, i.r.Logger, param, i.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	ctrl := controller.InitializeInfoCommandController(ctx, param, i.r.Runtime)
	return ctrl.Info(ctx, i.r.Logger.Logger, param) //nolint:wrapcheck
}
