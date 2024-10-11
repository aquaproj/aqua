package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

type updateAquaCommand struct {
	r *util.Param
}

func newUpdateAqua(r *util.Param) *cli.Command {
	i := &updateAquaCommand{
		r: r,
	}
	return &cli.Command{
		Action: i.action,
		Name:   "update-aqua",
		Aliases: []string{
			"upa",
		},
		Usage: "Update aqua",
		Description: `Update aqua.

e.g.
$ aqua update-aqua [version]

aqua is installed in $AQUA_ROOT_DIR/bin.
By default the latest version of aqua is installed, but you can specify the version with argument.

e.g.
$ aqua update-aqua # Install the latest version
$ aqua update-aqua v1.20.0 # Install v1.20.0
`,
	}
}

func (ua *updateAquaCommand) action(c *cli.Context) error {
	tracer, err := startTrace(c.String("trace"))
	if err != nil {
		return err
	}
	defer tracer.Stop()

	cpuProfiler, err := startCPUProfile(c.String("cpu-profile"))
	if err != nil {
		return err
	}
	defer cpuProfiler.Stop()

	param := &config.Param{}
	if err := setParam(c, ua.r.LogE, "update-aqua", param, ua.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl, err := controller.InitializeUpdateAquaCommandController(c.Context, param, http.DefaultClient, ua.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize a UpdateAquaController: %w", err)
	}
	return ctrl.UpdateAqua(c.Context, ua.r.LogE, param) //nolint:wrapcheck
}
