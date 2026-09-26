// Package fixcmd implements the aqua fix command.
package fixcmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/aquaproj/aqua/v2/pkg/controller/fixcmd"
	"github.com/urfave/cli/v3"
)

const description = `Bring aqua.yaml up to date with the registry.

	$ aqua fix

A configuration goes out of date without being edited, because what it refers to moves.
A package's repository is renamed or transferred, the registry keeps the old name as an
alias, and the file goes on working while naming something nobody uses -- until the
registry stops carrying the alias. This is what brings it up to date.

Nothing it does changes what the file means: the packages installed before and after a
run are the same packages at the same versions. That is what makes it safe to run on
every commit, in a hook or in the job that pushes fixes back to a pull request.

Only the standard registry's packages. Another registry's aliases are its own, and are
read from it at install time rather than from the table this uses.

aqua-lock.json follows by itself. It records a package under the name aqua.yaml asks for,
so the entry under the old name is one 'aqua lock update' drops and the one under the new
name is one it writes.

	$ aqua fix && aqua lock update

--check writes nothing and exits non-zero when there was something to do, which is how a
CI job asks whether the configuration still says what the registry says.

	$ aqua fix --check

--offline works from the table of other names already on disk instead of fetching one.
What can't be answered from it is left alone, so the result is a subset of what an online
run would write rather than something else.

	$ aqua fix --offline

What is not here is anything that changes what is installed, which is 'aqua up', and
anything that changes the shape of a configuration for a new aqua, which is a migration.
Neither is something a hook should do on its own.
`

type args struct {
	*cliargs.GlobalArgs

	Check   bool
	Offline bool
}

type command struct {
	r *util.Param
}

// New creates the "aqua fix" command.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	a := &args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:        "fix",
		Usage:       "Bring aqua.yaml up to date with the registry",
		Description: description,
		Action: func(ctx context.Context, _ *cli.Command) error {
			return i.action(ctx, a)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "check",
				Usage:       "Write nothing and exit non-zero if the configuration is out of date",
				Destination: &a.Check,
			},
			&cli.BoolFlag{
				Name:        "offline",
				Usage:       "Work from what is already on disk instead of fetching",
				Destination: &a.Offline,
			},
		},
	}
}

func (i *command) action(ctx context.Context, a *args) error {
	profiler, err := profile.Start(a.Trace, a.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	logger := i.r.Logger

	param := &config.Param{}
	if err := util.SetParam(a.GlobalArgs, logger, param, i.r.Version); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl, err := controller.InitializeFixCommandController(ctx, logger.Logger, param, &http.Client{})
	if err != nil {
		return fmt.Errorf("initialize a FixController: %w", err)
	}
	return ctrl.Fix(ctx, logger.Logger, param, &fixcmd.Args{ //nolint:wrapcheck
		Check:   a.Check,
		Offline: a.Offline,
	})
}
