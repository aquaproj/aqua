package genr

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
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
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "the maximum number of versions",
			},
			&cli.BoolFlag{
				Name:  "deep",
				Usage: "This flag was deprecated and had no meaning from aqua v2.15.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/2351",
			},
		},
	}
}

func (i *command) action(c *cli.Context) error {
	profiler, err := profile.Start(c)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(c, i.r.LogE, "generate-registry", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeGenerateRegistryCommandController(c.Context, param, http.DefaultClient, os.Stdout)
	return ctrl.GenerateRegistry(c.Context, param, i.r.LogE, c.Args().Slice()...) //nolint:wrapcheck
}
