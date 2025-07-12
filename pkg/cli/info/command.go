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

type infoCommand struct {
	r *util.Param
}

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

func (i *infoCommand) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	err := util.SetParam(cmd, i.r.LogE, "info", param, i.r.LDFlags)
	if err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl := controller.InitializeInfoCommandController(ctx, param, i.r.Runtime)

	return ctrl.Info(ctx, i.r.LogE, param) //nolint:wrapcheck
}
