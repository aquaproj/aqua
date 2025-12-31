// Package install implements the aqua install command for downloading and installing tools.
// The install command downloads packages according to configuration files,
// creates symbolic links, and manages package installations with various filtering options.
package install

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// Args holds command-line arguments for the install command.
type Args struct {
	*cliargs.GlobalArgs

	OnlyLink    bool
	Test        bool
	All         bool
	Tags        string
	ExcludeTags string
}

// command holds the parameters and configuration for the install command.
type command struct {
	r    *util.Param
	args *Args
}

// New creates and returns a new CLI command for installing tools.
// The returned command handles tool installation with options for
// link-only mode, tag filtering, and global configuration inclusion.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r:    r,
		args: args,
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
				Name:        "only-link",
				Aliases:     []string{"l"},
				Usage:       "create links but skip downloading packages",
				Destination: &args.OnlyLink,
			},
			&cli.BoolFlag{
				Name:        "test",
				Usage:       "This flag was deprecated and had no meaning from aqua v2.0.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/1691",
				Destination: &args.Test,
			},
			&cli.BoolFlag{
				Name:        "all",
				Aliases:     []string{"a"},
				Usage:       "install all aqua configuration packages",
				Destination: &args.All,
			},
			&cli.StringFlag{
				Name:        "tags",
				Aliases:     []string{"t"},
				Usage:       "filter installed packages with tags",
				Destination: &args.Tags,
			},
			&cli.StringFlag{
				Name:        "exclude-tags",
				Usage:       "exclude installed packages with tags",
				Destination: &args.ExcludeTags,
			},
		},
	}
}

// action implements the main logic for the install command.
// It initializes the install controller and executes the tool installation
// process based on configuration files and command line options.
func (i *command) action(ctx context.Context, _ *cli.Command) error {
	profiler, err := profile.Start(i.args.Trace, i.args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(i.args.GlobalArgs, i.r.Logger, param, i.r.Version); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	param.OnlyLink = i.args.OnlyLink
	param.All = i.args.All
	param.Tags = parseTags(i.args.Tags)
	param.ExcludedTags = parseTags(i.args.ExcludeTags)

	ctrl, err := controller.InitializeInstallCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize an InstallController: %w", err)
	}
	return ctrl.Install(ctx, i.r.Logger.Logger, param) //nolint:wrapcheck
}

func parseTags(tags string) map[string]struct{} {
	tagsM := map[string]struct{}{}
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		tagsM[tag] = struct{}{}
	}
	return tagsM
}
