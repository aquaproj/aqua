package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"time"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *LDFlags
	LogE    *logrus.Entry
	Runtime *runtime.Runtime
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (runner *Runner) setParam(c *cli.Context, param *config.Param) error {
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
	param.RootDir = config.GetRootDir(osenv.New())
	logE := runner.LogE
	log.SetLevel(param.LogLevel, logE)
	param.MaxParallelism = config.GetMaxParallelism(os.Getenv("AQUA_MAX_PARALLELISM"), logE)
	param.GlobalConfigFilePaths = finder.ParseGlobalConfigFilePaths(os.Getenv("AQUA_GLOBAL_CONFIG"))
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	param.PWD = wd
	return nil
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
			&cli.StringFlag{
				Name:  "trace",
				Usage: "trace output file path",
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

type Tracer struct {
	f io.Closer
}

func (t *Tracer) Stop() {
	if t == nil {
		return
	}
	trace.Stop()
}

func startTrace(p string) (*Tracer, error) {
	if p == "" {
		return nil, nil //nolint:nilnil
	}
	f, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("create a trace output file: %w", err)
	}
	if err := trace.Start(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("start a trace: %w", err)
	}
	return &Tracer{
		f: f,
	}, nil
}
