package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

const generateRegistryDescription = `Generate a Registry's package configuration.

$ aqua gr cli/cli > registry.yaml

You can also insert a package configuration into the existing configuration file with -i option.

$ aqua gr -i registry.yaml cli/cli
`

func (runner *Runner) newScaffoldCommand() *cli.Command {
	return &cli.Command{
		Name:        "generate-registry",
		Aliases:     []string{"gr"},
		Usage:       "Generate a registry's package configuration",
		ArgsUsage:   `<package name>`,
		Description: generateRegistryDescription,
		Action:      runner.generateRegistryAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "i",
				Usage: "Insert a registry to configuration file",
			},
		},
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
	ctrl := controller.InitializeScaffoldCommandController(c.Context, param, http.DefaultClient)
	return ctrl.Scaffold(c.Context, param, runner.LogE, c.Args().Slice()...) //nolint:wrapcheck
}
