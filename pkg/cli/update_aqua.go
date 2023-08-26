package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newUpdateAquaCommand() *cli.Command {
	return &cli.Command{
		Name:  "update-aqua",
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
		Action: r.updaetAquaAction,
	}
}

func (r *Runner) updaetAquaAction(c *cli.Context) error {
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
	if err := r.setParam(c, "update-aqua", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeUpdateAquaCommandController(c.Context, param, http.DefaultClient, r.Runtime)
	return ctrl.UpdateAqua(c.Context, r.LogE, param) //nolint:wrapcheck
}
