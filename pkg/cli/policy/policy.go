package policy

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v3"
)

func New(r *util.Param) *cli.Command {
	return &cli.Command{
		Name:  "policy",
		Usage: "Manage Policy",
		Subcommands: []*cli.Command{
			newPolicyAllow(r),
			newPolicyDeny(r),
			newPolicyInit(r),
		},
	}
}
