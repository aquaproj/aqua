package updateaqua

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

type updateAquaCommand struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
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

func (ua *updateAquaCommand) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, ua.r.LogE, "update-aqua", param, ua.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl, err := controller.InitializeUpdateAquaCommandController(ctx, ua.r.LogE, param, http.DefaultClient, ua.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize an UpdateAquaController: %w", err)
	}
	return ctrl.UpdateAqua(ctx, ua.r.LogE, param) //nolint:wrapcheck
}
