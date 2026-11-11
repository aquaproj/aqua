package migrate

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/aquaproj/aqua/v2/pkg/controller/migrate"
	"github.com/urfave/cli/v3"
)

const checksumDescription = `Take out what aqua-checksums.json was for.

	$ aqua migrate checksum

A lock file records the checksum of every environment of every package a configuration asks
for, so a repository with one says the same thing twice: aqua-checksums.json is the same
answer, kept where install used to look for it. What the settings in aqua.yaml did -- turn
the file on, insist every package be in it -- a lock file does by being the thing install
reads.

So the settings come out and the file goes. supported_envs stays: it says which
environments a lock file records, and that is still a question.

It is refused for a configuration whose lock file doesn't hold every package it asks for.
The settings coming out is the checksum file stopping being read, and doing that while
something isn't in the lock file would leave that package installed against nothing.

	$ aqua lock update && aqua migrate checksum

--keep-file leaves aqua-checksums.json where it is. Somebody else on the repository may be
on an aqua that still reads it, and this command can't know; the settings come out either
way, because this aqua doesn't read them.

	$ aqua migrate checksum --keep-file

--check writes nothing and exits non-zero when there is something to take out, which is how
a job asks whether this has been run.

	$ aqua migrate checksum --check
`

type checksumArgs struct {
	*cliargs.GlobalArgs

	KeepFile bool
	Check    bool
}

type checksumCommand struct {
	r *util.Param
}

func newChecksum(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &checksumArgs{
		GlobalArgs: globalArgs,
	}
	i := &checksumCommand{
		r: r,
	}
	return &cli.Command{
		Name:        "checksum",
		Usage:       "Take out what aqua-checksums.json was for",
		Description: checksumDescription,
		Action: func(ctx context.Context, _ *cli.Command) error {
			return i.action(ctx, args)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "keep-file",
				Usage:       "Leave aqua-checksums.json where it is",
				Destination: &args.KeepFile,
			},
			&cli.BoolFlag{
				Name:        "check",
				Usage:       "Write nothing and exit non-zero if there is something to take out",
				Destination: &args.Check,
			},
		},
	}
}

func (i *checksumCommand) action(ctx context.Context, args *checksumArgs) error {
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

	ctrl, err := controller.InitializeMigrateCommandController(ctx, logger.Logger, param)
	if err != nil {
		return fmt.Errorf("initialize a MigrateController: %w", err)
	}
	return ctrl.Checksum(logger.Logger, param.CWD, param.ConfigFilePath, &migrate.ChecksumArgs{ //nolint:wrapcheck
		Check:    args.Check,
		KeepFile: args.KeepFile,
	})
}
