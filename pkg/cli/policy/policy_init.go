package policy

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// policyInitArgs holds command-line arguments for the policy init command.
type policyInitArgs struct {
	*cliargs.GlobalArgs

	FilePath []string
}

// policyInitCommand holds the parameters and configuration for the policy init command.
type policyInitCommand struct {
	r    *util.Param
	args *policyInitArgs
}

// newPolicyInit creates and returns a new CLI command for initializing policy files.
// The returned command creates new policy files with default configuration
// to establish security policies for package installations.
func newPolicyInit(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &policyInitArgs{
		GlobalArgs: globalArgs,
	}
	i := &policyInitCommand{
		r:    r,
		args: args,
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
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "file_path",
				Min:         0,
				Max:         1,
				Destination: &args.FilePath,
			},
		},
	}
}

// action implements the main logic for the policy init command.
// It initializes the policy init controller and creates a new policy file
// at the specified path with default security policy configuration.
func (pi *policyInitCommand) action(ctx context.Context, _ *cli.Command) error {
	profiler, err := profile.Start(pi.args.Trace, pi.args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(pi.args.GlobalArgs, pi.r.Logger, param, pi.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	ctrl := controller.InitializeInitPolicyCommandController(ctx)
	filePath := ""
	if len(pi.args.FilePath) > 0 {
		filePath = pi.args.FilePath[0]
	}
	return ctrl.Init(pi.r.Logger.Logger, filePath) //nolint:wrapcheck
}
