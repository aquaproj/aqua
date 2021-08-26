package cli

import (
	"context"
	"io"
	"time"

	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *LDFlags
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (runner *Runner) Run(ctx context.Context, args ...string) error {
	compiledDate, err := time.Parse(time.RFC3339, runner.LDFlags.Date)
	if err != nil {
		compiledDate = time.Now()
	}
	app := cli.App{
		Name:     "aqua",
		Usage:    "General version manager. https://github.com/suzuki-shunsuke/aqua",
		Version:  runner.LDFlags.Version + " (" + runner.LDFlags.Commit + ")",
		Compiled: compiledDate,
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
		Commands: []*cli.Command{
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install tools",
				Action:  runner.installAction,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "only-link",
						Usage: "create links but skip download packages",
					},
				},
			},
			{
				Name:   "exec",
				Usage:  "Execute tool",
				Action: runner.execAction,
			},
			{
				Name:   "get-bin-dir",
				Usage:  "Get the configuration `bin_dir`",
				Action: runner.getBinDirAction,
			},
			{
				Name:   "version",
				Usage:  "Show version",
				Action: runner.versionAction,
			},
		},
	}

	return app.RunContext(ctx, args) //nolint:wrapcheck
}
