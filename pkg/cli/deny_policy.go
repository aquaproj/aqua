package cli //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newDenyPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:  "disallow",
		Usage: "Deny a policy file",
		Description: `Deny a policy file
e.g.
$ aqua policy disallow [<policy file path>]
`,
		Action: runner.disallowPolicyAction,
	}
}

func (runner *Runner) disallowPolicyAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "disallow-policy", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeDenyPolicyCommandController(c.Context, param)
	return ctrl.Deny(c.Context, runner.LogE, param, c.Args().First()) //nolint:wrapcheck
}
