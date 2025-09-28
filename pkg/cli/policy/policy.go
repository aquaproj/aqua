// Package policy implements the aqua policy commands for managing security policies.
// The policy commands provide functionality to configure, allow, and deny packages
// based on security policies to ensure safe package installations.
package policy

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v3"
)

// New creates and returns a new CLI command for policy management.
// The returned command provides subcommands for allowing, denying,
// and initializing security policies for package installations.
func New(r *util.Param) *cli.Command {
	return &cli.Command{
		Name:  "policy",
		Usage: "Manage Policy",
		Commands: []*cli.Command{
			newPolicyAllow(r),
			newPolicyDeny(r),
			newPolicyInit(r),
		},
	}
}
