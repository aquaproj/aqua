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

func (runner *Runner) Run(ctx context.Context, args ...string) error { //nolint:funlen
	compiledDate, err := time.Parse(time.RFC3339, runner.LDFlags.Date)
	if err != nil {
		compiledDate = time.Now()
	}
	app := cli.App{
		Name:           "aqua",
		Usage:          "Version Manager of CLI. https://github.com/aquaproj/aqua",
		Version:        runner.LDFlags.Version + " (" + runner.LDFlags.Commit + ")",
		Compiled:       compiledDate,
		ExitErrHandler: exitErrHandlerFunc,
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
						Name:    "only-link",
						Aliases: []string{"l"},
						Usage:   "create links but skip download packages",
					},
					&cli.BoolFlag{
						Name:  "test",
						Usage: "test file.src after installing the package",
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "install all aqua configuration packages",
					},
				},
			},
			{
				Name:   "exec",
				Usage:  "Execute tool",
				Action: runner.execAction,
			},
			{
				Name:   "init",
				Usage:  "Create aqua.yaml",
				Action: runner.initAction,
			},
			{
				Name:   "list",
				Usage:  "List packages in Registries",
				Action: runner.listAction,
			},
			{
				Name:   "which",
				Usage:  "Output the file path of the given command",
				Action: runner.whichAction,
			},
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "Search packages in registries and output the configuration interactively",
				Action:  runner.generateAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "f",
						Usage: "the file path of packages list.",
					},
				},
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

func exitErrHandlerFunc(c *cli.Context, err error) {
	if c.Command.Name != "exec" {
		cli.HandleExitCoder(err)
		return
	}
}
