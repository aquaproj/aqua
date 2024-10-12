package version

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v2"
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

func (i *command) action(c *cli.Context) error {
	cli.ShowVersion(c)
	return nil
}
