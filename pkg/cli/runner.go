package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/aquaproj/aqua/pkg/policy"
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

func (runner *Runner) setParam(c *cli.Context, commandName string, param *config.Param) error {
	param.Args = c.Args().Slice()
	if logLevel := c.String("log-level"); logLevel != "" {
		param.LogLevel = logLevel
	}
	param.ConfigFilePath = c.String("config")
	param.Dest = c.String("o")
	param.OnlyLink = c.Bool("only-link")
	param.IsTest = c.Bool("test")
	if commandName == "generate-registry" {
		param.InsertFile = c.String("i")
	} else {
		param.Insert = c.Bool("i")
	}
	param.All = c.Bool("all")
	param.Prune = c.Bool("prune")
	param.SelectVersion = c.Bool("select-version")
	param.File = c.String("f")
	param.LogColor = os.Getenv("AQUA_LOG_COLOR")
	param.AQUAVersion = runner.LDFlags.Version
	param.RootDir = config.GetRootDir(osenv.New())
	homeDir, _ := os.UserHomeDir()
	param.HomeDir = homeDir
	logE := runner.LogE
	log.SetLevel(param.LogLevel, logE)
	log.SetColor(param.LogColor, logE)
	param.MaxParallelism = config.GetMaxParallelism(os.Getenv("AQUA_MAX_PARALLELISM"), logE)
	param.GlobalConfigFilePaths = finder.ParseGlobalConfigFilePaths(os.Getenv("AQUA_GLOBAL_CONFIG"))
	param.Deep = c.Bool("deep")
	param.Pin = c.Bool("pin")
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	param.PWD = wd
	param.ProgressBar = os.Getenv("AQUA_PROGRESS_BAR") == "true"
	param.PolicyConfigFilePaths = policy.ParseEnv(os.Getenv("AQUA_POLICY_CONFIG"))
	param.Tags = parseTags(strings.Split(c.String("tags"), ","))
	param.ExcludedTags = parseTags(strings.Split(c.String("exclude-tags"), ","))
	return nil
}

func parseTags(tags []string) map[string]struct{} {
	tagsM := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		tagsM[tag] = struct{}{}
	}
	return tagsM
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
			&cli.StringFlag{
				Name:  "cpu-profile",
				Usage: "cpu profile output file path",
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			runner.newInitCommand(),
			runner.newInitPolicyCommand(),
			runner.newInstallCommand(),
			runner.newUpdateAquaCommand(),
			runner.newGenerateCommand(),
			runner.newWhichCommand(),
			runner.newExecCommand(),
			runner.newListCommand(),
			runner.newGenerateRegistryCommand(),
			runner.newCompletionCommand(),
			runner.newVersionCommand(),
			runner.newCpCommand(),
			runner.newUpdateChecksumCommand(),
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
