package lock

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/aquaproj/aqua/v2/pkg/controller/lockupdate"
	"github.com/urfave/cli/v3"
)

const updateDescription = `Create or update aqua-lock.json.

	$ aqua lock update

aqua-lock.json sits beside aqua.yaml and records how each package in it is actually
installed: the asset name, the format, the files inside the archive, the signature
settings and the checksum, for every OS and architecture the package supports.
aqua installs from the lock file rather than from a registry, so what is installed
is what was reviewed when the file was committed. Commit it together with aqua.yaml.

Packages already in the lock file are left alone, because keeping the recorded answer
is the point of the file. Use --force to resolve them again, which is what you want
after editing a registry.

	$ aqua lock update --force

Given package names, only those packages are updated. A name may carry a version, in
the same "<package name>@<version>" form aqua.yaml uses.

	$ aqua lock update cli/cli
	$ aqua lock update --force cli/cli@v2.70.0
`

type updateArgs struct {
	*cliargs.GlobalArgs

	Force    bool
	Packages []string
}

type updateCommand struct {
	r *util.Param
}

func newUpdate(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &updateArgs{
		GlobalArgs: globalArgs,
	}
	i := &updateCommand{
		r: r,
	}
	return &cli.Command{
		Name:        "update",
		Usage:       "Create or update the lock file",
		ArgsUsage:   `[<package name>[@<version>] ...]`,
		Description: updateDescription,
		Action: func(ctx context.Context, _ *cli.Command) error {
			return i.action(ctx, args)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Usage:       "Update packages which are already in the lock file",
				Destination: &args.Force,
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "packages",
				Min:         0,
				Max:         -1,
				Destination: &args.Packages,
			},
		},
	}
}

func (i *updateCommand) action(ctx context.Context, args *updateArgs) error {
	profiler, err := profile.Start(args.Trace, args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	logger := i.r.Logger

	param := &config.Param{}
	if err := util.SetParam(args.GlobalArgs, logger, param, i.r.Version); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl, err := controller.InitializeLockUpdateCommandController(ctx, logger.Logger, param, &http.Client{})
	if err != nil {
		return fmt.Errorf("initialize a LockUpdateController: %w", err)
	}
	return ctrl.Update(ctx, logger.Logger, param, &lockupdate.Args{ //nolint:wrapcheck
		Force:    args.Force,
		Packages: args.Packages,
	})
}
