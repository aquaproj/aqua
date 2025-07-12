package policy

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

type policyInitCommand struct {
	r *util.Param
}

func newPolicyInit(r *util.Param) *cli.Command {
	i := &policyInitCommand{
		r: r,
	}

	return &cli.Command{
		Action:    i.action,
		Name:      "init",
		Usage:     "Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua-policy.yaml">]`,
		Description: `Create a policy file if it doesn't exist
e.g.
$ aqua policy init # create "aqua-policy.yaml"
$ aqua policy init foo.yaml # create foo.yaml`,
	}
}

func (pi *policyInitCommand) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	err := util.SetParam(cmd, pi.r.LogE, "init-policy", param, pi.r.LDFlags)
	if err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl := controller.InitializeInitPolicyCommandController(ctx)

	return ctrl.Init(pi.r.LogE, cmd.Args().First()) //nolint:wrapcheck
}
