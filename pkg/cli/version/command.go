package version

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v3"
)

type command struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:   "version",
		Usage:  "Show version",
		Action: i.action,
	}
}

func (i *command) action(context.Context, *cli.Command) error {
	fmt.Fprintln(i.r.Stdout, "aqua version "+i.r.LDFlags.GetVersion())
	return nil
}
