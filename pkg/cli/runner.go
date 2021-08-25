package cli

import (
	"context"
	"io"

	"github.com/suzuki-shunsuke/aqua/pkg/constant"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (runner *Runner) Run(ctx context.Context, args ...string) error {
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
						EnvVars: []string{"AQUA_LOG_LEVEL"},
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "configuration file path",
						EnvVars: []string{"AQUA_CONFIG"},
					},
				},
			},
			{
				Name:   "exec",
				Usage:  "Execute tool",
				Action: runner.execAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "log-level",
						Usage:   "log level",
						EnvVars: []string{"AQUA_LOG_LEVEL"},
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "configuration file path",
						EnvVars: []string{"AQUA_CONFIG"},
					},
				},
			},
		},
	}

	return app.RunContext(ctx, args) //nolint:wrapcheck
}
