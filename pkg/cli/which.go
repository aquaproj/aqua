package cli

import (
	"fmt"

	"github.com/suzuki-shunsuke/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) whichAction(c *cli.Context) error {
	param := &controller.Param{}
	if err := runner.setCLIArg(c, param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl, err := controller.New(c.Context, param)
	if err != nil {
		return fmt.Errorf("initialize a controller: %w", err)
	}

	exeName, _, err := parseExecArgs(c.Args().Slice())
	if err != nil {
		return err
	}

	return ctrl.Which(c.Context, param, exeName) //nolint:wrapcheck
}
