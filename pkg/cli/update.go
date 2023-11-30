package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:        "update",
		Aliases:     []string{"up"},
		Usage:       "Update registries and packages",
		Description: updateDescription,
		Action:      r.updateAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "i",
				Usage: `Select packages with fuzzy finder`,
			},
			&cli.BoolFlag{
				Name:    "select-version",
				Aliases: []string{"s"},
				Usage:   `Select the version with fuzzy finder. Default to display 30 versions, use --limit/-l to change it.`,
			},
			&cli.BoolFlag{
				Name:    "only-registry",
				Aliases: []string{"r"},
				Usage:   `Update only registries`,
			},
			&cli.BoolFlag{
				Name:    "only-package",
				Aliases: []string{"p"},
				Usage:   `Update only packages`,
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "The maximum number of versions. Non-positive number refers to no limit.",
				Value:   config.DefaultVerCnt,
			},
		},
	}
}

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
If this command finds a aqua.yaml, it ignores other aqua.yaml including global configuration files ($AQUA_GLOBAL_CONFIG).

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
`

func (r *Runner) updateAction(c *cli.Context) error {
	tracer, err := startTrace(c.String("trace"))
	if err != nil {
		return err
	}
	defer tracer.Stop()

	cpuProfiler, err := startCPUProfile(c.String("cpu-profile"))
	if err != nil {
		return err
	}
	defer cpuProfiler.Stop()

	param := &config.Param{}
	if err := r.setParam(c, "update", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeUpdateCommandController(c.Context, param, http.DefaultClient, r.Runtime)
	return ctrl.Update(c.Context, r.LogE, param) //nolint:wrapcheck
}
