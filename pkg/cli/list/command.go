// Package list implements the aqua list command for listing packages in registries.
// The list command displays available packages from configured registries,
// with options to show only installed packages and include global configurations.
package list

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// command holds the parameters and configuration for the list command.
type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for listing packages.
// The returned command allows users to list available packages from registries
// with optional filtering for installed packages and global configurations.
func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
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
				Name:    "installed",
				Aliases: []string{"i"},
				Usage:   "List installed packages",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "List global configuration packages too",
			},
		},
	}
}

// action implements the main logic for the list command.
// It initializes the list controller and executes the package listing
// operation based on the provided command line flags and configuration.
func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, i.r.Logger, "list", param, i.r.Version); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeListCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	return ctrl.List(ctx, i.r.Logger.Logger, param) //nolint:wrapcheck
}
