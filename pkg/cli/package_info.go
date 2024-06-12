package cli //nolint:dupl

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newPackageInfoCommand() *cli.Command {
	return &cli.Command{
		Name: "package",
		Usage: "Show package defintions",
		Description: `Show definition of packages.
e.g.
$ aqua info package sigstore/cosign`,
		Action: r.packageInfoAction,
	}
}

func (r *Runner) packageInfoAction(c *cli.Context) error {
	param := &config.Param{}
	if err := r.setParam(c, "generate", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl := controller.InitializePackageInfoCommandController(c.Context, param, http.DefaultClient, r.Runtime)
	return ctrl.PackageInfo(c.Context, param, r.LogE, c.Args().Slice()...)
}
