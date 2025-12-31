// Package upc implements the aqua update-checksum command for updating package checksums.
// The update-checksum command updates the checksums of packages in configuration files,
// ensuring package integrity and security after version updates.
package upc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// Args holds command-line arguments for the update-checksum command.
type Args struct {
	*cliargs.GlobalArgs

	All   bool
	Deep  bool
	Prune bool
}

// command holds the parameters and configuration for the update-checksum command.
type command struct {
	r    *util.Param
	args *Args
}

// New creates and returns a new CLI command for updating package checksums.
// The returned command provides functionality to update checksums in
// configuration files after package version changes.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r:    r,
		args: args,
	}
	return &cli.Command{
		Name: "update-checksum",
		Aliases: []string{
			"upc",
		},
		Usage: "Create or Update aqua-checksums.json",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "all",
				Aliases:     []string{"a"},
				Usage:       "Create or Update all aqua-checksums.json including global configuration",
				Destination: &args.All,
			},
			&cli.BoolFlag{
				Name:        "deep",
				Usage:       "This flag was deprecated and had no meaning from aqua v2.0.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/1769",
				Destination: &args.Deep,
			},
			&cli.BoolFlag{
				Name:        "prune",
				Usage:       "Remove unused checksums",
				Destination: &args.Prune,
			},
		},
		Description: `Create or Update aqua-checksums.json.

e.g.
$ aqua update-checksum

By default aqua doesn't update aqua-checksums.json of the global configuration.
If you want to update them too,
please set "-a" option.

$ aqua update-checksum -a

By default, aqua update-checksum doesn't remove existing checksums even if they aren't unused.
If -prune option is set, aqua unused checksums would be removed.

$ aqua update-checksum -prune
`,
		Action: i.action,
	}
}

func (i *command) action(ctx context.Context, _ *cli.Command) error {
	profiler, err := profile.Start(i.args.Trace, i.args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(i.args.GlobalArgs, i.r.Logger, param, i.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	param.All = i.args.All
	param.Prune = i.args.Prune
	ctrl := controller.InitializeUpdateChecksumCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	return ctrl.UpdateChecksum(ctx, i.r.Logger.Logger, param) //nolint:wrapcheck
}
