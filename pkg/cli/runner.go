package cli

import (
	"context"
	"io"
	"time"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/sirupsen/logrus"
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

func (runner *Runner) setParam(c *cli.Context, param *config.Param) (*logrus.Entry, error) { //nolint:unparam
	if logLevel := c.String("log-level"); logLevel != "" {
		param.LogLevel = logLevel
	}
	param.ConfigFilePath = c.String("config")
	param.OnlyLink = c.Bool("only-link")
	param.IsTest = c.Bool("test")
	param.Insert = c.Bool("i")
	param.All = c.Bool("all")
	param.File = c.String("f")
	param.AQUAVersion = runner.LDFlags.Version
	param.RootDir = config.GetRootDir()
	logE := log.New(param.AQUAVersion)
	log.SetLevel(param.LogLevel, logE)
	param.MaxParallelism = config.GetMaxParallelism(logE)
	return logE, nil
}

func (runner *Runner) Run(ctx context.Context, args ...string) error {
	compiledDate, err := time.Parse(time.RFC3339, runner.LDFlags.Date)
	if err != nil {
		compiledDate = time.Now()
	}
	app := cli.App{
		Name:           "aqua",
		Usage:          "Version Manager of CLI. https://aquaproj.github.io/",
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
			runner.newInstallCommand(),
			runner.newExecCommand(),
			runner.newInitCommand(),
			runner.newListCommand(),
			runner.newWhichCommand(),
			runner.newGenerateCommand(),
			runner.newVersionCommand(),
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
