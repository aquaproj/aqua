package vacuum

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/urfave/cli/v2"
)

const description = `Perform vacuuming tasks.

Enable vacuuming by setting the AQUA_VACUUM_DAYS environment variable to a value greater than 0.
This command removes versions of packages that have not been used for the specified number of days.

You can list all packages managed by the vacuum system or only expired packages.

	# Show all packages managed by the vacuum system
	$ aqua vacuum show

	# Show only expired packages
	$ aqua vacuum show --expired
	$ aqua vacuum show -e

	# Run vacuum cleaning
	$ aqua vacuum run
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
		Subcommands: []*cli.Command{
			{
				Name:    "show",
				Aliases: []string{"s"},
				Usage:   "Show packages managed by vacuum system",
				Action:  i.action,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "expired",
						Usage:   "Show only expired packages",
						Aliases: []string{"e"},
					},
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run vacuum cleaning",
				Action:  i.action,
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

	if param.VacuumDays == 0 {
		return errors.New("vacuum is not enabled, please set the AQUA_VACUUM_DAYS environment variable")
	}

	ctrl := controller.InitializeVacuumCommandController(c.Context, param, http.DefaultClient, i.r.Runtime)

	defer func() {
		if err := ctrl.Close(logE); err != nil {
			logerr.WithError(logE, err).Error("close vacuum controller")
		}
	}()

	if c.Command.Name == "show" {
		if err := ctrl.ListPackages(c.Context, logE, c.Bool("expired")); err != nil {
			return fmt.Errorf("show packages: %w", err)
		}
		return nil
	}

	if c.Command.Name == "run" {
		if err := ctrl.Vacuum(c.Context, logE); err != nil {
			return fmt.Errorf("run: %w", err)
		}
	}

	return nil
}
