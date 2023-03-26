package cli //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newAllowPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:  "allow-policy",
		Usage: "Allow a policy file",
		Description: `Allow a policy file
e.g.
$ aqua allow-policy [<policy file path>]
`,
		Action: runner.allowPolicyAction,
	}
}

func (runner *Runner) allowPolicyAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "allow-policy", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeAllowPolicyCommandController(c.Context, param)
	return ctrl.Allow(c.Context, runner.LogE, param, c.Args().First()) //nolint:wrapcheck
}
