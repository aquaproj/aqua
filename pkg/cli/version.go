package cli

import (
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newVersionCommand() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "Show version",
		Action: runner.versionAction,
	}
}

func (runner *Runner) versionAction(c *cli.Context) error {
	cli.ShowVersion(c)
	return nil
}
