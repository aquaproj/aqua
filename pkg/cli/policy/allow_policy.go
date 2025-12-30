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

// policyAllowCommand holds the parameters and configuration for the policy allow command.
type policyAllowCommand struct {
	r *util.Param
}

// newPolicyAllow creates and returns a new CLI command for allowing policy files.
// The returned command marks a policy file as allowed, permitting packages
// to be installed according to that policy.
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

// action implements the main logic for the policy allow command.
// It initializes the allow policy controller and marks the specified
// policy file as allowed based on the provided file path.
func (pa *policyAllowCommand) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, pa.r.Logger, "allow-policy", param, pa.r.Version); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeAllowPolicyCommandController(ctx, param)
	return ctrl.Allow(pa.r.Logger.Logger, param, cmd.Args().First()) //nolint:wrapcheck
}
