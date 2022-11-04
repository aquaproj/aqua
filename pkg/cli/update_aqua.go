package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newUpdateAquaCommand() *cli.Command {
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
		Action: runner.updaetAquaAction,
	}
}

func (runner *Runner) updaetAquaAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "update-aqua", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeUpdateAquaCommandController(c.Context, param, http.DefaultClient, runner.Runtime)
	return ctrl.UpdateAqua(c.Context, runner.LogE, param) //nolint:wrapcheck
}
