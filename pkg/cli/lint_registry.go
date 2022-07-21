package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newLintRegistryCommand() *cli.Command {
	return &cli.Command{
		Name: "lint-registry",
		Aliases: []string{
			"lr",
		},
		Usage:     "Lint a registry file",
		ArgsUsage: `<registry file path>`,
		Description: `Lint a registry file
e.g.
$ aqua lr registry.yaml`,
		Action: runner.lintRegistryAction,
	}
}

func (runner *Runner) lintRegistryAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "lint-registry", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInitCommandController(c.Context, param)
	return ctrl.Init(c.Context, c.Args().First(), runner.LogE) //nolint:wrapcheck
}
