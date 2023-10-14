package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update tools",
		Description: `Update tools.

e.g.
$ aqua update
`,
		Action: r.updateAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "i",
				Usage: `Select packages with fuzzy finder`,
			},
			&cli.BoolFlag{
				Name:    "only-registry",
				Aliases: []string{"r"},
				Usage:   `Update only registries`,
			},
			&cli.BoolFlag{
				Name:    "only-package",
				Aliases: []string{"p"},
				Usage:   `Update only packages`,
			},
		},
	}
}

func (r *Runner) updateAction(c *cli.Context) error {
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
	if err := r.setParam(c, "update", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeUpdateCommandController(c.Context, param, http.DefaultClient, r.Runtime)
	return ctrl.Update(c.Context, r.LogE, param) //nolint:wrapcheck
}
