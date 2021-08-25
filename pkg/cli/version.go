package cli

import (
	"github.com/urfave/cli/v2"
)

func (runner *Runner) versionAction(c *cli.Context) error {
	cli.ShowVersion(c)
	return nil
}
