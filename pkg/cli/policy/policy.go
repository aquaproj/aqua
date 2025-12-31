// Package policy implements the aqua policy commands for managing security policies.
// The policy commands provide functionality to configure, allow, and deny packages
// based on security policies to ensure safe package installations.
package policy

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v3"
)

// New creates and returns a new CLI command for policy management.
// The returned command provides subcommands for allowing, denying,
// and initializing security policies for package installations.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	return &cli.Command{
		Name:  "policy",
		Usage: "Manage Policy",
		Commands: []*cli.Command{
			newPolicyAllow(r, globalArgs),
			newPolicyDeny(r, globalArgs),
			newPolicyInit(r, globalArgs),
		},
	}
}

// NewInitPolicy creates a top-level policy-init command for backward compatibility.
// This is a deprecated command that creates policy files. Users should use
// "aqua policy init" command instead for new implementations.
func NewInitPolicy(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &policyInitArgs{
		GlobalArgs: globalArgs,
	}
	policyInit := &policyInitCommand{
		r: r,
	}
	return &cli.Command{
		Name:      "init-policy",
		Usage:     "[Deprecated] Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua-policy.yaml">]`,
		Description: `[Deprecated] Create a policy file if it doesn't exist

Please use "aqua policy init" command instead.

e.g.
$ aqua init-policy # create "aqua-policy.yaml"
$ aqua init-policy foo.yaml # create foo.yaml`,
		Action: func(ctx context.Context, _ *cli.Command) error {
			return policyInit.action(ctx, args)
		},
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
