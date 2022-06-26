package cli

import (
	"fmt"
	"net/http"

	"github.com/clivm/clivm/pkg/config"
	"github.com/clivm/clivm/pkg/controller"
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
`

func (runner *Runner) newGenerateRegistryCommand() *cli.Command {
	return &cli.Command{
		Name:        "generate-registry",
		Aliases:     []string{"gr"},
		Usage:       "Generate a registry's package configuration",
		ArgsUsage:   `<package name>`,
		Description: generateRegistryDescription,
		Action:      runner.generateRegistryAction,
		// TODO support "i" option
		// Flags: []cli.Flag{
		// 	&cli.StringFlag{
		// 		Name:  "i",
		// 		Usage: "Insert a registry to configuration file",
		// 	},
		// },
	}
}

func (runner *Runner) generateRegistryAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "generate-registry", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeGenerateRegistryCommandController(c.Context, param, http.DefaultClient)
	return ctrl.GenerateRegistry(c.Context, param, runner.LogE, c.Args().Slice()...) //nolint:wrapcheck
}
