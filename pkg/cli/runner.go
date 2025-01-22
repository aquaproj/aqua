package cli

import (
	"context"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/cli/completion"
	"github.com/aquaproj/aqua/v2/pkg/cli/cp"
	"github.com/aquaproj/aqua/v2/pkg/cli/exec"
	"github.com/aquaproj/aqua/v2/pkg/cli/generate"
	"github.com/aquaproj/aqua/v2/pkg/cli/genr"
	"github.com/aquaproj/aqua/v2/pkg/cli/info"
	"github.com/aquaproj/aqua/v2/pkg/cli/initcmd"
	"github.com/aquaproj/aqua/v2/pkg/cli/install"
	"github.com/aquaproj/aqua/v2/pkg/cli/list"
	cpolicy "github.com/aquaproj/aqua/v2/pkg/cli/policy"
	"github.com/aquaproj/aqua/v2/pkg/cli/remove"
	"github.com/aquaproj/aqua/v2/pkg/cli/root"
	"github.com/aquaproj/aqua/v2/pkg/cli/upc"
	"github.com/aquaproj/aqua/v2/pkg/cli/update"
	"github.com/aquaproj/aqua/v2/pkg/cli/updateaqua"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/cli/vacuum"
	"github.com/aquaproj/aqua/v2/pkg/cli/version"
	"github.com/aquaproj/aqua/v2/pkg/cli/which"
	"github.com/urfave/cli/v2"
)

type Runner struct{}

type newC func(r *util.Param) *cli.Command

func commands(param *util.Param, newCs ...newC) []*cli.Command {
	cs := make([]*cli.Command, len(newCs))
	for i, newC := range newCs {
		cs[i] = newC(param)
	}
	return cs
}

func Run(ctx context.Context, param *util.Param, args ...string) error { //nolint:funlen
	compiledDate, err := time.Parse(time.RFC3339, param.LDFlags.Date)
	if err != nil {
		compiledDate = time.Now()
	}
	app := cli.App{
		Name:           "aqua",
		Usage:          "Version Manager of CLI. https://aquaproj.github.io/",
		Version:        param.LDFlags.Version + " (" + param.LDFlags.Commit + ")",
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
			&cli.BoolFlag{
				Name:    "disable-cosign",
				Usage:   "Disable Cosign verification",
				EnvVars: []string{"AQUA_DISABLE_COSIGN"},
			},
			&cli.BoolFlag{
				Name:    "disable-slsa",
				Usage:   "Disable SLSA verification",
				EnvVars: []string{"AQUA_DISABLE_SLSA"},
			},
			&cli.BoolFlag{
				Name:    "disable-github-artifact-attestation",
				Usage:   "Disable GitHub Artifact Attestations verification",
				EnvVars: []string{"AQUA_DISABLE_GITHUB_ARTIFACT_ATTESTATION"},
			},
			&cli.StringFlag{
				Name:  "trace",
				Usage: "trace output file path",
			},
			&cli.StringFlag{
				Name:  "cpu-profile",
				Usage: "cpu profile output file path",
			},
			&cli.IntFlag{
				Name:    "vacuum-days",
				Usage:   "Vacuum days",
				EnvVars: []string{"AQUA_VACUUM_DAYS"},
				Value:   60, //nolint:mnd
			},
		},
		EnableBashCompletion: true,
		Commands: commands(
			param,
			info.New,
			initcmd.New,
			cpolicy.New,
			cpolicy.NewInitPolicy,
			install.New,
			updateaqua.New,
			generate.New,
			which.New,
			exec.New,
			list.New,
			genr.New,
			completion.New,
			version.New,
			cp.New,
			root.New,
			upc.New,
			remove.New,
			update.New,
			vacuum.New,
		),
	}

	return app.RunContext(ctx, args) //nolint:wrapcheck
}

func exitErrHandlerFunc(c *cli.Context, err error) {
	if c.Command.Name != "exec" {
		cli.HandleExitCoder(err)
		return
	}
}
