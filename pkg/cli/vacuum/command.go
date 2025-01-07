package vacuum

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

const description = `Perform vacuuming tasks.
If no argument is provided, the vacuum will clean expired packages.

	# Execute vacuum cleaning
	$ aqua vacuum

This command has an alias "v".

	$ aqua v

Enable vacuuming by setting the AQUA_VACUUM_DAYS environment variable to a value greater than 0.
This command removes versions of packages that have not been used for the specified number of days.

You can list all packages managed by the vacuum system or only expired packages.

	# List all packages managed by the vacuum system
	$ aqua vacuum --list
	$ aqua vacuum -l

	# List only expired packages
	$ aqua vacuum --expired
	$ aqua vacuum -e

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
				Name:     "list",
				Usage:    "List all packages managed by vacuum system",
				Category: "list",
			},
			&cli.BoolFlag{
				Name:     "expired",
				Usage:    "List only expired packages",
				Category: "list",
			},
		},
	}
}

// Define the vacuum modes for the CLI
const (
	ListPackages          string = "list-packages"
	ListExpiredPackages   string = "list-expired-packages"
	VacuumExpiredPackages string = "vacuum-expired-packages"
)

func (i *command) action(c *cli.Context) error {
	profiler, err := profile.Start(c)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	mode := parseVacuumMode(c)

	param := &config.Param{}
	if err := util.SetParam(c, i.r.LogE, "vacuum", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	if param.VacuumDays == nil {
		return errors.New("vacuum is not enabled, please set the AQUA_VACUUM_DAYS environment variable")
	}

	ctrl := controller.InitializeVacuumCommandController(c.Context, param, http.DefaultClient, i.r.Runtime)

	vacuumMode, err := ctrl.GetVacuumModeCLI(mode)
	if err != nil {
		return fmt.Errorf("get vacuum mode: %w", err)
	}

	if err := ctrl.Vacuum(c.Context, i.r.LogE, vacuumMode, nil); err != nil {
		return fmt.Errorf("vacuum: %w", err)
	}
	return nil
}

func parseVacuumMode(c *cli.Context) string {
	mode := VacuumExpiredPackages
	if c.Bool("list") {
		mode = ListPackages
	}
	if c.Bool("expired") {
		mode = ListExpiredPackages
	}
	return mode
}
