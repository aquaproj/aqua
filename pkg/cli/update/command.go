// Package update implements the aqua update command for updating packages and registries.
// The update command updates packages and registries to their latest versions,
// fetching version information from various sources like GitHub Releases and Tags.
package update

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

const updateDescription = `Update registries and packages.
If no argument is passed, all registries and packages are updated to the latest.

  # Update all packages and registries to the latest versions
  $ aqua update

This command has an alias "up"

  $ aqua up

This command gets the latest version from GitHub Releases, GitHub Tags, and crates.io and updates aqua.yaml.
This command doesn't update commit hashes.
This command doesn't install packages.
This command updates only a nearest aqua.yaml from the current directory.
If this command finds an aqua.yaml, it ignores other aqua.yaml including global configuration files ($AQUA_GLOBAL_CONFIG).

So if you want to update other files, please change the current directory or specify the configuration file path with the option '-c'.

  $ aqua -c foo/aqua.yaml update

If you want to update only registries, please use the --only-registry [-r] option.

  # Update only registries
  $ aqua update -r

If you want to update only packages, please use the --only-package [-p] option.

  # Update only packages
  $ aqua update -p

If you want to update only specific packages, please use the -i option.
You can select packages with the fuzzy finder.
If -i option is used, registries aren't updated.

  # Select updated packages with fuzzy finder
  $ aqua update -i

If you want to select versions, please use the --select-version [-s] option.
You can select versions with the fuzzy finder. You can not only update but also downgrade packages.
By default, -s will only display 30 versions of package.
Use --limit/-l to change it. Non-positive number refers to no limit.

  # Select updated packages and versions with fuzzy finder
  # Display 30 versions by default
  $ aqua update -i -s
  # Display only 5 versions
  $ aqua update -i -s -l 5
  # Display all versions
  $ aqua update -i -s -l -1

This command doesn't update packages if the field 'version' is used.

  packages:
    - name: cli/cli@v2.0.0 # Update
    - name: gohugoio/hugo
      version: v0.118.0 # Doesn't update

So if you don't want to update specific packages, the field 'version' is useful.

You can specify packages with command names. aqua finds packages that have these commands and updates them.

  $ aqua update <command name> [<command name> ...]

e.g.

  # Update cli/cli
  $ aqua update gh

You can also specify a version.

  $ aqua update gh@v2.30.0

You can also filter updated packages using package tags.

e.g.
$ aqua up -t foo # Install only packages having a tag "foo"
$ aqua up --exclude-tags foo # Install only packages not having a tag "foo"
`

// Args holds command-line arguments for the update command.
type Args struct {
	*cliargs.GlobalArgs

	Interactive   bool
	SelectVersion bool
	OnlyRegistry  bool
	OnlyPackage   bool
	Limit         int
	Tags          string
	ExcludeTags   string
	Packages      []string
}

type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for updating packages and registries.
// The returned command provides functionality to update packages to their
// latest versions from various sources like GitHub and crates.io.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:        "update",
		Aliases:     []string{"up"},
		Usage:       "Update registries and packages",
		Description: updateDescription,
		Action: func(ctx context.Context, _ *cli.Command) error {
			return i.action(ctx, args)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "i",
				Usage:       `Select packages with fuzzy finder`,
				Destination: &args.Interactive,
			},
			&cli.BoolFlag{
				Name:        "select-version",
				Aliases:     []string{"s"},
				Usage:       `Select the version with fuzzy finder. Default to display 30 versions, use --limit/-l to change it.`,
				Destination: &args.SelectVersion,
			},
			&cli.BoolFlag{
				Name:        "only-registry",
				Aliases:     []string{"r"},
				Usage:       `Update only registries`,
				Destination: &args.OnlyRegistry,
			},
			&cli.BoolFlag{
				Name:        "only-package",
				Aliases:     []string{"p"},
				Usage:       `Update only packages`,
				Destination: &args.OnlyPackage,
			},
			&cli.IntFlag{
				Name:        "limit",
				Aliases:     []string{"l"},
				Usage:       "The maximum number of versions. Non-positive number refers to no limit.",
				Value:       config.DefaultVerCnt,
				Destination: &args.Limit,
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
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "packages",
				Destination: &args.Packages,
				Max:         -1,
			},
		},
	}
}

// action implements the main logic for the update command.
// It initializes the update controller and executes package and registry
// updates based on the provided command line arguments.
func (i *command) action(ctx context.Context, args *Args) error {
	profiler, err := profile.Start(args.Trace, args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(args.GlobalArgs, i.r.Logger, param, i.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	param.Insert = args.Interactive
	param.SelectVersion = args.SelectVersion
	param.OnlyPackage = args.OnlyPackage
	param.OnlyRegistry = args.OnlyRegistry
	param.Limit = args.Limit
	param.Tags = util.ParseTags(strings.Split(args.Tags, ","))
	param.ExcludedTags = util.ParseTags(strings.Split(args.ExcludeTags, ","))
	param.Args = args.Packages
	ctrl := controller.InitializeUpdateCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	return ctrl.Update(ctx, i.r.Logger.Logger, param) //nolint:wrapcheck
}
