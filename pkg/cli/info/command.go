// Package info implements the aqua info command for displaying system information.
// The info command shows various details about the aqua installation,
// configuration, and environment for debugging and troubleshooting purposes.
package info

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// infoCommand holds the parameters and configuration for the info command.
type infoCommand struct {
	r *util.Param
}

// New creates and returns a new CLI command for displaying information.
// The returned command shows system information about aqua installation,
// configuration paths, and environment details for troubleshooting.
func New(r *util.Param) *cli.Command {
	i := &infoCommand{
		r: r,
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
func (i *infoCommand) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, i.r.LogE, "info", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInfoCommandController(ctx, param, i.r.Runtime)
	return ctrl.Info(ctx, i.r.LogE, param) //nolint:wrapcheck
}
