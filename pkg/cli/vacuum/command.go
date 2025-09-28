// Package vacuum implements the aqua vacuum command for cleaning up unused packages.
// The vacuum command removes unused installed packages based on usage timestamps,
// helping users save storage space and maintain a clean installation directory.
package vacuum

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

const description = `Remove unused installed packages.

This command removes unused installed packages, which is useful to save storage and keep your machine clean.

	$ aqua vacuum

It removes installed packages which haven't been used for over the expiration days.
The default expiration days is 60, but you can change it by the environment variable $AQUA_VACUUM_DAYS or the command line option "-days <expiration days>".

e.g.

	$ export AQUA_VACUUM_DAYS=90

	$ aqua vacuum -d 30

As of aqua v2.43.0, aqua records packages' last used date times.
Date times are updated when packages are installed or executed.
Packages installed by aqua v2.42.2 or older don't have records of last used date times, so aqua can't remove them.
To solve the problem, "aqua vacuum --init" is available.

	aqua vacuum --init

"aqua vacuum --init" searches installed packages from aqua.yaml including $AQUA_GLOBAL_CONFIG and records the current date time as the last used date time of those packages if their last used date times aren't recorded.

"aqua vacuum --init" can't record date times of install packages which are not found in aqua.yaml.
If you want to record their date times, you need to remove them by "aqua rm" command and re-install them.
`

type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for cleaning up unused packages.
// The returned command provides functionality to remove packages that haven't
// been used for a specified number of days.
func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:        "vacuum",
		Usage:       "Remove unused installed packages",
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
				Usage:   "Expiration days",
				Sources: cli.EnvVars("AQUA_VACUUM_DAYS"),
				Value:   60, //nolint:mnd
			},
		},
	}
}

// action implements the main logic for the vacuum command.
// It initializes the vacuum controller and removes unused packages
// based on the expiration days configuration.
func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	logE := i.r.LogE

	param := &config.Param{}
	if err := util.SetParam(cmd, logE, "vacuum", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	if cmd.Bool("init") {
		ctrl := controller.InitializeVacuumInitCommandController(ctx, i.r.LogE, param, i.r.Runtime, &http.Client{})
		if err := ctrl.Init(ctx, logE, param); err != nil {
			return err //nolint:wrapcheck
		}
		return nil
	}

	param.VacuumDays = cmd.Int("days")
	if param.VacuumDays <= 0 {
		return errors.New("vacuum days must be greater than 0")
	}

	ctrl := controller.InitializeVacuumCommandController(ctx, param, i.r.Runtime)
	if err := ctrl.Vacuum(logE, param); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}
