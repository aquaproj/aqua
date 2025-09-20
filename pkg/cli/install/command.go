// Package install implements the aqua install command for downloading and installing tools.
// The install command downloads packages according to configuration files,
// creates symbolic links, and manages package installations with various filtering options.
package install

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

// command holds the parameters and configuration for the install command.
type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for installing tools.
// The returned command handles tool installation with options for
// link-only mode, tag filtering, and global configuration inclusion.
func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "Install tools",
		Description: `Install tools according to the configuration files.

e.g.
$ aqua i

If you want to create only symbolic links and want to skip downloading package, please set "-l" option.

$ aqua i -l

By default aqua doesn't install packages in the global configuration.
If you want to install packages in the global configuration too,
please set "-a" option.

$ aqua i -a

You can filter installed packages with package tags.

e.g.
$ aqua i -t foo # Install only packages having a tag "foo"
$ aqua i --exclude-tags foo # Install only packages not having a tag "foo"
`,
		Action: i.action,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "only-link",
				Aliases: []string{"l"},
				Usage:   "create links but skip downloading packages",
			},
			&cli.BoolFlag{
				Name:  "test",
				Usage: "This flag was deprecated and had no meaning from aqua v2.0.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/1691",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "install all aqua configuration packages",
			},
			&cli.StringFlag{
				Name:    "tags",
				Aliases: []string{"t"},
				Usage:   "filter installed packages with tags",
			},
			&cli.StringFlag{
				Name:  "exclude-tags",
				Usage: "exclude installed packages with tags",
			},
		},
	}
}

// action implements the main logic for the install command.
// It initializes the install controller and executes the tool installation
// process based on configuration files and command line options.
func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, i.r.LogE, "install", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl, err := controller.InitializeInstallCommandController(ctx, i.r.LogE, param, http.DefaultClient, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize an InstallController: %w", err)
	}
	return ctrl.Install(ctx, i.r.LogE, param) //nolint:wrapcheck
}
