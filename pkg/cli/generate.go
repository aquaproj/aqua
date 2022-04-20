package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/aquaproj/aqua/pkg/log"
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

You can pass packages with positional arguments.

$ aqua g [<registry name>,<package name> ...]

$ aqua g standard,cli/cli standard,junegunn/fzf
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0

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
- name: cli/cli@v2.2.0`

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
		},
	}
}

func (runner *Runner) generateAction(c *cli.Context) error {
	param := &config.Param{}
	if err := runner.setCLIArg(c, param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	logE := log.New(param.AQUAVersion)
	log.SetLevel(param.LogLevel, logE)

	ctrl := controller.InitializeGenerateCommandController(c.Context, param.AQUAVersion, param)

	return ctrl.Generate(c.Context, logE, param, c.Args().Slice()...) //nolint:wrapcheck
}
