// Package generate implements the aqua generate command for interactive package configuration.
// The generate command allows users to search for packages in registries and generate
// configuration entries interactively, supporting both command-line and fuzzy finder interfaces.
package generate

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

// command holds the parameters and configuration for the generate command.
type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for package configuration generation.
// The returned command provides interactive package search and configuration
// generation capabilities with various output and selection options.
func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:        "generate",
		Aliases:     []string{"g"},
		Usage:       "Search packages in registries and output the configuration interactively",
		ArgsUsage:   `[<registry name>,<package name> ...]`,
		Description: generateDescription,
		Action:      i.action,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "f",
				Usage: `the file path of packages list. When the value is "-", the list is passed from the standard input`,
			},
			&cli.BoolFlag{
				Name:  "i",
				Usage: `Insert packages to configuration file`,
			},
			&cli.BoolFlag{
				Name:  "pin",
				Usage: `Pin version`,
			},
			&cli.BoolFlag{
				Name:  "g",
				Usage: `Insert packages in a global configuration file`,
			},
			&cli.BoolFlag{
				Name:    "detail",
				Aliases: []string{"d"},
				Usage:   `Output additional fields such as description and link`,
				Sources: cli.EnvVars("AQUA_GENERATE_WITH_DETAIL"),
			},
			&cli.StringFlag{
				Name:  "o",
				Usage: `inserted file`,
			},
			&cli.BoolFlag{
				Name:    "select-version",
				Aliases: []string{"s"},
				Usage:   `Select the installed version interactively. Default to display 30 versions, use --limit/-l to change it.`,
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

// action implements the main logic for the generate command.
// It initializes the generate controller and executes the package search
// and configuration generation process based on user input.
func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, i.r.Logger, "generate", param, i.r.Version); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeGenerateCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	return ctrl.Generate(ctx, i.r.Logger.Logger, param, cmd.Args().Slice()...) //nolint:wrapcheck
}

const generateDescription = `Search packages in registries and output the configuration interactively.

If no argument is passed, interactive fuzzy finder is launched.

$ aqua g

  influxdata/influx-cli (standard) (influx)                     ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
  newrelic/newrelic-cli (standard) (newrelic)                   │  cli/cli
  pivotal-cf/pivnet-cli (standard) (pivnet)                     │
  scaleway/scaleway-cli (standard) (scw)                        │  https://cli.github.com/
  tfmigrator/cli (standard) (tfmigrator)                        │  GitHub’cs official command line tool
  aws/copilot-cli (standard) (copilot)                          │
  codeclimate/test-reporter (standard)                          │
  create-go-app/cli (standard) (cgapp)                          │
  harness/drone-cli (standard) (drone)                          │
  sigstore/rekor (standard) (rekor-cli)                         │
  getsentry/sentry-cli (standard)                               │
  knative/client (standard) (kn)                                │
  rancher/cli (standard) (rancher)                              │
  tektoncd/cli (standard) (tkn)                                 │
  civo/cli (standard) (civo)                                    │
  dapr/cli (standard) (dapr)                                    │
  mongodb/mongocli (standard)                                   │
  openfaas/faas-cli (standard)                                  │
> cli/cli (standard) (gh)                                       │
  48/380                                                        │
> cli                                                           └ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─

Please select the package you want to install, then the package configuration is outputted.
You can select multiple packages by tab key.
Please copy and paste the outputted configuration in the aqua configuration file.

$ aqua g # tfmigrator/cli is selected
- name: tfmigrator/cli@v0.2.1

You can update the configuration file directly with "-i" option.

$ aqua g -i

You can update an imported file with "-o" option.

$ aqua g -o aqua/pkgs.yaml

You can pass packages with positional arguments.

$ aqua g [<registry name>,<package name>[@<version>] ...]

$ aqua g standard,cli/cli standard,junegunn/fzf standard,suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0
- name: suzuki-shunsuke/tfcmt@v3.0.0

You can omit the registry name if it is "standard".

$ aqua g cli/cli
- name: cli/cli@v2.2.0

With "-f" option, you can pass packages.

$ aqua g -f packages.txt # list of <registry name>,<package name>
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0
- name: tfmigrator/cli@v0.2.1

$ cat packages.txt | aqua g -f -
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0
- name: tfmigrator/cli@v0.2.1

$ aqua list | aqua g -f - # Generate configuration to install all packages

You can omit the registry name if it is "standard".

echo "cli/cli" | aqua g -f -
- name: cli/cli@v2.2.0

You can select a version interactively with "-s" option.
By default, aqua g -s will only display 30 versions of package.
Use --limit/-l to change it. Non-positive number refers to no limit.

# Display 30 versions of selected by default
$ aqua g -s
# Display all versions of selected package
$ aqua g -s -l -1
# Display 5 versions of selected package
$ aqua g -s -l 5

The option "-pin" is useful to prevent the package from being updated by Renovate.

$ aqua g -pin cli/cli
- name: cli/cli
  version: v2.2.0

With -detail option, aqua outputs additional information such as description and link.

$ aqua g -detail cli/cli
- name: cli/cli@v2.2.0
  description: GitHub’s official command line tool
  link: https://github.com/cli/cli

With -g option, aqua reads a first global configuration file.

$ aqua g -g cli/cli

You can add packages to a first global configuration file with -g and -i option.

$ aqua g -g -i cli/cli
`
