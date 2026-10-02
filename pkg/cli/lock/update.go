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

Each package version is cached once it has been resolved, and the cache never expires:
what it holds describes one version of one package, which doesn't change. --no-cache
refetches and rewrites it, for the case where a cached file was written wrong.

	$ aqua lock update --no-cache cli/cli@v2.70.0

An entry for a package aqua.yaml no longer asks for is removed. What a lock file is for
is the packages a configuration asks for, and an entry for one it doesn't answers a
question nobody puts. Naming packages leaves the rest alone: such a run isn't about
aqua.yaml, and every package it didn't name would go.
`

type updateArgs struct {
	*cliargs.GlobalArgs

	Force    bool
	NoCache  bool
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
			&cli.BoolFlag{
				Name:        "no-cache",
				Usage:       "Ignore the cache and fetch registries again",
				Destination: &args.NoCache,
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
	param.NoCache = args.NoCache

	ctrl, err := controller.InitializeLockUpdateCommandController(ctx, logger.Logger, param, &http.Client{}, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize a LockUpdateController: %w", err)
	}
	return ctrl.Update(ctx, logger.Logger, param, &lockupdate.Args{ //nolint:wrapcheck
		Force:    args.Force,
		Packages: args.Packages,
	})
}
