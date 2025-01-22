package vacuum

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

const description = `Perform vacuuming tasks.

Enable vacuuming by setting the AQUA_VACUUM_DAYS environment variable to a value greater than 0.
This command removes versions of packages that have not been used for the specified number of days.

You can list all packages managed by the vacuum system or only expired packages.

	$ aqua vacuum
`

type command struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:        "vacuum",
		Usage:       "Operate vacuuming tasks (If AQUA_VACUUM_DAYS is set)",
		Aliases:     []string{"v"},
		Description: description,
		Action:      i.action,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "init",
				Usage: "Create timestamp files.",
			},
			&cli.IntFlag{
				Name:    "days",
				Aliases: []string{"d"},
				Usage:   "Vacuum days",
				EnvVars: []string{"AQUA_VACUUM_DAYS"},
				Value:   60, //nolint:mnd
			},
		},
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
	if err := util.SetParam(c, logE, "vacuum", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	if c.Bool("init") {
		ctrl := controller.InitializeVacuumInitCommandController(c.Context, param, i.r.Runtime, &http.Client{})
		if err := ctrl.Init(c.Context, logE, param); err != nil {
			return err //nolint:wrapcheck
		}
		return nil
	}

	ctrl := controller.InitializeVacuumCommandController(c.Context, param, i.r.Runtime)
	if err := ctrl.Vacuum(logE, param); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}
