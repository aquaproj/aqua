package policy //nolint:dupl

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

type policyAllowCommand struct {
	r *util.Param
}

func newPolicyAllow(r *util.Param) *cli.Command {
	i := &policyAllowCommand{
		r: r,
	}

	return &cli.Command{
		Action: i.action,
		Name:   "allow",
		Usage:  "Allow a policy file",
		Description: `Allow a policy file
e.g.
$ aqua policy allow [<policy file path>]
`,
	}
}

func (pa *policyAllowCommand) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	err := util.SetParam(cmd, pa.r.LogE, "allow-policy", param, pa.r.LDFlags)
	if err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl := controller.InitializeAllowPolicyCommandController(ctx, param)

	return ctrl.Allow(pa.r.LogE, param, cmd.Args().First()) //nolint:wrapcheck
}
