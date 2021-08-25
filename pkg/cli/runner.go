package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/suzuki-shunsuke/aqua/pkg/constant"
	"github.com/suzuki-shunsuke/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (runner *Runner) Run(ctx context.Context, args ...string) error {
	if len(args) != 0 {
		exeName := filepath.Base(args[0])
		if exeName != "aqua" {
			param := &controller.Param{}
			ctrl, err := controller.New(ctx, param)
			if err != nil {
				return fmt.Errorf("initialize a controller: %w", err)
			}
			return ctrl.Exec(ctx, param, exeName, args[1:]) //nolint:wrapcheck
		}
	}
	app := cli.App{
		Name:    "aqua",
		Usage:   "General version manager. https://github.com/suzuki-shunsuke/aqua",
		Version: constant.Version,
		Commands: []*cli.Command{
			{
				Name:   "add",
				Usage:  "Install tools",
				Action: runner.installAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "log-level",
						Usage:   "log level",
						EnvVars: []string{"CUBE_LOG_LEVEL"},
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "configuration file path",
						EnvVars: []string{"CUBE_CONFIG"},
					},
				},
			},
		},
	}

	return app.RunContext(ctx, args) //nolint:wrapcheck
}
