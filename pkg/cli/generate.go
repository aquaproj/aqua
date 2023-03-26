package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

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

Please select the package you want to install, then the package configuration is outptted.
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

$ aqua g -s

The option "-pin" is useful to prevent the package from being updated by Renovate.

$ aqua g -pin cli/cli
- name: cli/cli
  version: v2.2.0
`

func (runner *Runner) newGenerateCommand() *cli.Command {
	return &cli.Command{
		Name:        "generate",
		Aliases:     []string{"g"},
		Usage:       "Search packages in registries and output the configuration interactively",
		ArgsUsage:   `[<registry name>,<package name> ...]`,
		Description: generateDescription,
		Action:      runner.generateAction,
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
			&cli.StringFlag{
				Name:  "o",
				Usage: `inserted file`,
			},
			&cli.BoolFlag{
				Name:    "select-version",
				Aliases: []string{"s"},
				Usage:   `Select the installed version interactively`,
			},
		},
	}
}

func (runner *Runner) generateAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "generate", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeGenerateCommandController(c.Context, param, http.DefaultClient, runner.Runtime)
	return ctrl.Generate(c.Context, runner.LogE, param, c.Args().Slice()...) //nolint:wrapcheck
}
