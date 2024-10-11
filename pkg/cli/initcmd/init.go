package initcmd

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cpuprofile"
	"github.com/aquaproj/aqua/v2/pkg/cli/tracer"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

type initCommand struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
	ic := &initCommand{
		r: r,
	}
	return &cli.Command{
		Name:      "init",
		Usage:     "Create a configuration file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua.yaml">]`,
		Description: `Create a configuration file if it doesn't exist
e.g.
$ aqua init # create "aqua.yaml"
$ aqua init foo.yaml # create foo.yaml`,
		Action: ic.action,
		Flags:  []cli.Flag{},
	}
}

func (ic *initCommand) action(c *cli.Context) error {
	tracer, err := tracer.Start(c.String("trace"))
	if err != nil {
		return err
	}
	defer tracer.Stop()

	cpuProfiler, err := cpuprofile.Start(c.String("cpu-profile"))
	if err != nil {
		return err
	}
	defer cpuProfiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(c, ic.r.LogE, "init", param, ic.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInitCommandController(c.Context, param)
	return ctrl.Init(c.Context, ic.r.LogE, c.Args().First()) //nolint:wrapcheck
}
