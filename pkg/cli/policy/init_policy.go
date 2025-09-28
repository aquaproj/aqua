package policy

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v3"
)

// NewInitPolicy creates and returns a new CLI command for initializing policy files.
// This is a deprecated command that creates policy files. Users should use
// "aqua policy init" command instead for new implementations.
func NewInitPolicy(r *util.Param) *cli.Command {
	cmd := newPolicyInit(r)
	return &cli.Command{
		Name:      "init-policy",
		Usage:     "[Deprecated] Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua-policy.yaml">]`,
		Description: `[Deprecated] Create a policy file if it doesn't exist

Please use "aqua policy init" command instead.

e.g.
$ aqua init-policy # create "aqua-policy.yaml"
$ aqua init-policy foo.yaml # create foo.yaml`,
		Action: cmd.Action,
	}
}
