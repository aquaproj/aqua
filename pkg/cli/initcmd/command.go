package initcmd

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/aquaproj/aqua/v2/pkg/controller/initcmd"
	"github.com/urfave/cli/v3"
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
$ aqua init foo.yaml # create foo.yaml
$ aqua init -u # Replace "packages:" with "import_dir: imports"
$ aqua init -i <directory path> # Replace "packages:" with "import_dir: <directory path>"
$ aqua init -d # Create a directory "aqua" and create "aqua/aqua.yaml"
`,
		Action: ic.action,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "use-import-dir",
				Aliases: []string{"u"},
				Usage:   "Use import_dir",
			},
			&cli.StringFlag{
				Name:    "import-dir",
				Aliases: []string{"i"},
				Usage:   "import_dir",
			},
			&cli.BoolFlag{
				Name:    "create-dir",
				Aliases: []string{"d"},
				Usage:   "Create a directory named aqua and create aqua.yaml in it",
			},
		},
	}
}

func (ic *initCommand) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, ic.r.LogE, "init", param, ic.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInitCommandController(ctx, param)
	cParam := &initcmd.Param{
		IsDir:     cmd.Bool("create-dir"),
		ImportDir: cmd.String("import-dir"),
	}
	if cParam.ImportDir == "" && cmd.Bool("use-import-dir") {
		cParam.ImportDir = "imports"
	}
	return ctrl.Init(ctx, ic.r.LogE, cmd.Args().First(), cParam) //nolint:wrapcheck
}
