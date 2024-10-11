package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

type infoCommand struct {
	r *Runner
}

func newInfo(r *Runner) *cli.Command {
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

func (i *infoCommand) action(c *cli.Context) error {
	tracer, err := startTrace(c.String("trace"))
	if err != nil {
		return err
	}
	defer tracer.Stop()

	cpuProfiler, err := startCPUProfile(c.String("cpu-profile"))
	if err != nil {
		return err
	}
	defer cpuProfiler.Stop()

	param := &config.Param{}
	if err := setParam(c, i.r.LogE, "info", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInfoCommandController(c.Context, param, i.r.Runtime)
	return ctrl.Info(c.Context, i.r.LogE, param) //nolint:wrapcheck
}
