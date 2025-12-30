// Package genr implements the aqua generate-registry command for creating registry configurations.
// The generate-registry command creates templates for registry package configurations,
// providing a starting point for adding new packages to the aqua registry.
package genr

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

const generateRegistryDescription = `Generate a template of Registry package configuration.

Note that you probably fix the generate code manually.
The generate code is not perfect and may include the wrong configuration.
It is just a template.

e.g.

$ aqua gr cli/cli # Outputs the configuration.
packages:
  - type: github_release
    repo_owner: cli
    repo_name: cli
    asset: gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}
    format: tar.gz
    description: GitHub’s official command line tool
    replacements:
      darwin: macOS
    overrides:
      - goos: windows
        format: zip
    supported_envs:
      - darwin
      - linux
      - amd64
    rosetta2: true

By default, aqua gets the information from the latest GitHub Releases.
You can specify a specific package version.

e.g.

$ aqua gr cli/cli@v2.0.0

By default, aqua gr gets all GitHub Releases to generate version_overrides.
You can limit the number of GitHub Releases by --limit.

e.g.

$ aqua gr --limit 100 suzuki-shunsuke/tfcmt

If --out-testdata is set, aqua inserts testdata into the specified file.

e.g.

$ aqua gr --out-testdata testdata.yaml suzuki-shunsuke/tfcmt

If -cmd is set, aqua sets files.

e.g.

$ aqua gr -cmd gh cli/cli

  files:
	  - name: gh

You can specify multiple commands with commas ",".

e.g.

$ aqua gr -cmd age,age-keygen FiloSottile/age

  files:
	  - name: age
	  - name: age-keygen
`

type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for generating registry configurations.
// The returned command provides functionality to generate template configurations
// for adding new packages to the aqua registry.
func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:        "generate-registry",
		Aliases:     []string{"gr"},
		Usage:       "Generate a registry's package configuration",
		ArgsUsage:   `<package name>`,
		Description: generateRegistryDescription,
		Action:      i.action,
		// TODO support "i" option
		Flags: []cli.Flag{
			// 	&cli.StringFlag{
			// 		Name:  "i",
			// 		Usage: "Insert a registry to configuration file",
			// 	},
			&cli.StringFlag{
				Name:  "out-testdata",
				Usage: "A file path where the testdata is outputted",
			},
			&cli.StringFlag{
				Name:  "cmd",
				Usage: "A list of commands joined with commas ','",
			},
			&cli.StringFlag{
				Name:    "generate-config",
				Aliases: []string{"c"},
				Usage:   "A configuration file path",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "the maximum number of versions",
			},
			&cli.BoolFlag{
				Name:  "deep",
				Usage: "This flag was deprecated and had no meaning from aqua v2.15.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/2351",
			},
			&cli.BoolFlag{
				Name:  "init",
				Usage: "Generate a configuration file",
			},
		},
	}
}

// action implements the main logic for the generate-registry command.
// It initializes the generate-registry controller and creates template
// configurations for new packages in the registry.
func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, i.r.Logger, "generate-registry", param, i.r.Version); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeGenerateRegistryCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, os.Stdout)
	return ctrl.GenerateRegistry(ctx, param, i.r.Logger.Logger, cmd.Args().Slice()...) //nolint:wrapcheck
}
