// Package list implements the aqua list command for listing packages in registries.
// The list command displays available packages from configured registries,
// with options to show only installed packages and include global configurations.
package list

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

// Args holds command-line arguments for the list command.
type Args struct {
	*cliargs.GlobalArgs

	Installed bool
	All       bool
}

// command holds the parameters and configuration for the list command.
type command struct {
	r    *util.Param
	args *Args
}

// New creates and returns a new CLI command for listing packages.
// The returned command allows users to list available packages from registries
// with optional filtering for installed packages and global configurations.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r:    r,
		args: args,
	}
	return &cli.Command{
		Name:   "list",
		Usage:  "List packages in Registries",
		Action: i.action,
		Description: `Output the list of packages in registries.
The output format is <registry name>,<package name>

e.g.
$ aqua list
standard,99designs/aws-vault
standard,abiosoft/colima
standard,abs-lang/abs
...

If the option -installed is set, the command lists only installed packages.

$ aqua list -installed
standard,golangci/golangci-lint,v1.56.2
standard,goreleaser/goreleaser,v1.24.0
...

By default, the command doesn't list global configuration packages.
If you want to list global configuration packages too, please set the option -a.

$ aqua list -installed -a
`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "installed",
				Aliases:     []string{"i"},
				Usage:       "List installed packages",
				Destination: &args.Installed,
			},
			&cli.BoolFlag{
				Name:        "all",
				Aliases:     []string{"a"},
				Usage:       "List global configuration packages too",
				Destination: &args.All,
			},
		},
	}
}

// action implements the main logic for the list command.
// It initializes the list controller and executes the package listing
// operation based on the provided command line flags and configuration.
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
	param.Installed = i.args.Installed
	param.All = i.args.All
	ctrl := controller.InitializeListCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	return ctrl.List(ctx, i.r.Logger.Logger, param) //nolint:wrapcheck
}
