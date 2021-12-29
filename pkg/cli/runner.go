package cli

import (
	"context"
	"io"
	"time"

	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *LDFlags
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (runner *Runner) Run(ctx context.Context, args ...string) error { //nolint:funlen
	compiledDate, err := time.Parse(time.RFC3339, runner.LDFlags.Date)
	if err != nil {
		compiledDate = time.Now()
	}
	app := cli.App{
		Name:           "aqua",
		Usage:          "Version Manager of CLI. https://github.com/aquaproj/aqua",
		Version:        runner.LDFlags.Version + " (" + runner.LDFlags.Commit + ")",
		Compiled:       compiledDate,
		ExitErrHandler: exitErrHandlerFunc,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level",
				EnvVars: []string{"AQUA_LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "configuration file path",
				EnvVars: []string{"AQUA_CONFIG"},
			},
		},
		Commands: []*cli.Command{
			{
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
`,
				Action: runner.installAction,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "only-link",
						Aliases: []string{"l"},
						Usage:   "create links but skip downloading packages",
					},
					&cli.BoolFlag{
						Name:  "test",
						Usage: "test file.src after installing the package",
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "install all aqua configuration packages",
					},
				},
			},
			{
				Name:  "exec",
				Usage: "Execute tool",
				Description: `Basically you don't have to use this command, because this is used by aqua internally. aqua-proxy invokes this command.
When you execute the command installed by aqua, "aqua exec" is executed internally.

e.g.
$ aqua exec -- gh version
gh version 2.4.0 (2021-12-21)
https://github.com/cli/cli/releases/tag/v2.4.0`,
				Action:    runner.execAction,
				ArgsUsage: `<executed command> [<arg> ...]`,
			},
			{
				Name:      "init",
				Usage:     "Create a configuration file if it doesn't exist",
				ArgsUsage: `[<created file path. The default value is "aqua.yaml">]`,
				Description: `Create a configuration file if it doesn't exist
e.g.
$ aqua init # create "aqua.yaml"
$ aqua init foo.yaml # create foo.yaml`,
				Action: runner.initAction,
			},
			{
				Name:   "list",
				Usage:  "List packages in Registries",
				Action: runner.listAction,
				Description: `Output the list of packages in registries.
The output format is <registry name>,<package name>

e.g.
$ aqua list
standard,99designs/aws-vault
standard,abiosoft/colima
standard,abs-lang/abs
...
`,
			},
			{
				Name:      "which",
				Usage:     "Output the absolute file path of the given command",
				ArgsUsage: `<command name>`,
				Description: `Output the absolute file path of the given command
e.g.
$ aqua which gh
/home/foo/.aqua/pkgs/github_release/github.com/cli/cli/v2.4.0/gh_2.4.0_macOS_amd64.tar.gz/gh_2.4.0_macOS_amd64/bin/gh

If the command isn't found in the configuration files, aqua searches the command in the environment variable PATH

$ aqua which ls
/bin/ls

If the command isn't found, exits with non zero exit code.

$ aqua which foo
FATA[0000] aqua failed                                   aqua_version=0.8.6 error="command is not found" exe_name=foo program=aqua
`,
				Action: runner.whichAction,
			},
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "Search packages in registries and output the configuration interactively",
				Description: `Search packages in registries and output the configuration interactively.
Interactive fuzzy finder is launched.

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
Please copy and paste the outputted configuration in the aqua configuration file.

$ aqua g # tfmigrator/cli is selected
- name: tfmigrator/cli@v0.2.1

You can update the configuration file directly by "aqua g >> <configuration file>".

$ aqua g >> aqua.yaml

With "-f" option, you can pass packages without interactive UI.

$ aqua g -f packages.txt # list of <registry name>,<package name>
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0
- name: tfmigrator/cli@v0.2.1

$ cat packages.txt | aqua g -f -
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0
- name: tfmigrator/cli@v0.2.1

You can omit the registry name if it is "standard".

echo "cli/cli" | aqua g -f -

$ aqua list | aqua g -f - # Generate configuration to install all packages`,
				Action: runner.generateAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "f",
						Usage: `the file path of packages list. When the value is "-", the list is passed from the standard input`,
					},
				},
			},
			{
				Name:   "version",
				Usage:  "Show version",
				Action: runner.versionAction,
			},
		},
	}

	return app.RunContext(ctx, args) //nolint:wrapcheck
}

func exitErrHandlerFunc(c *cli.Context, err error) {
	if c.Command.Name != "exec" {
		cli.HandleExitCoder(err)
		return
	}
}
