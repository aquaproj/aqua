package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

var errCommandIsRequired = errors.New("command is required")

func parseExecArgs(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, errCommandIsRequired
	}
	return filepath.Base(args[0]), args[1:], nil
}

func (runner *Runner) execAction(c *cli.Context) error {
	param := &controller.Param{}
	if err := runner.setCLIArg(c, param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl, err := controller.New(c.Context, param)
	if err != nil {
		return fmt.Errorf("initialize a controller: %w", err)
	}
	exeName, args, err := parseExecArgs(c.Args().Slice())
	if err != nil {
		return err
	}

	return ctrl.Exec(c.Context, param, exeName, args) //nolint:wrapcheck
}
