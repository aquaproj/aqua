package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newInfoCommand() *cli.Command {
	return &cli.Command{
		Name:  "info",
		Usage: "Show information",
		Description: `Show information.
e.g.
$ aqua info`,
		Action: runner.info,
	}
}

func (runner *Runner) info(c *cli.Context) error {
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
	if err := runner.setParam(c, "info", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInfoCommandController(c.Context, param, runner.Runtime)
	return ctrl.Info(c.Context, runner.LogE, param, c.Args().First()) //nolint:wrapcheck
}
