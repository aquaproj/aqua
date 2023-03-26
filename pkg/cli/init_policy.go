package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newInitPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:      "init-policy",
		Usage:     "[Deprecated] Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua-policy.yaml">]`,
		Description: `[Deprecated] Create a policy file if it doesn't exist

Please use "aqua policy set" command instead.

e.g.
$ aqua init-policy # create "aqua-policy.yaml"
$ aqua init-policy foo.yaml # create foo.yaml`,
		Action: runner.initPolicyAction,
	}
}

func (runner *Runner) initPolicyAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "init-policy", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInitPolicyCommandController(c.Context)
	return ctrl.Init(c.Context, c.Args().First(), runner.LogE) //nolint:wrapcheck
}
