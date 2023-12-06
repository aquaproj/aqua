package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newInitPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:      "init-policy",
		Usage:     "[Deprecated] Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua-policy.yaml">]`,
		Description: `[Deprecated] Create a policy file if it doesn't exist

Please use "aqua policy init" command instead.

e.g.
$ aqua init-policy # create "aqua-policy.yaml"
$ aqua init-policy foo.yaml # create foo.yaml`,
		Action: r.initPolicyAction,
	}
}

func (r *Runner) initPolicyAction(c *cli.Context) error {
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
	if err := r.setParam(c, "init-policy", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInitPolicyCommandController(c.Context)
	return ctrl.Init(r.LogE, c.Args().First()) //nolint:wrapcheck
}
