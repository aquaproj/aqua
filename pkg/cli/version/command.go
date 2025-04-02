package version

import (
	"context"

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

func (i *command) action(_ context.Context, cmd *cli.Command) error {
	cli.ShowVersion(cmd)
	return nil
}
