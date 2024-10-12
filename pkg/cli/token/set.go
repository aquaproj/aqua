package token

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/settoken"
	"github.com/urfave/cli/v2"
)

type setCommand struct {
	r *util.Param
}

func newSet(r *util.Param) *cli.Command {
	i := &setCommand{
		r: r,
	}
	return &cli.Command{
		Action:      i.action,
		Name:        "set",
		Usage:       "Set a GitHub access token in keyring",
		Description: `Set a GitHub access token in keyring`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "stdin",
				Usage: "Read a GitHub access token from stdin",
			},
		},
	}
}

func (pa *setCommand) action(c *cli.Context) error {
	profiler, err := profile.Start(c)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(c, pa.r.LogE, "token-set", param, pa.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := settoken.New(pa.r.Stdin, pa.r.Stdout)
	return ctrl.Set(c.Context, pa.r.LogE, param) //nolint:wrapcheck
}
