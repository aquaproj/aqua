package cli

import (
	"context"

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
	"github.com/aquaproj/aqua/v2/pkg/cli/which"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/urfave/cli/v3"
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

func Run(ctx context.Context, param *util.Param, args ...string) error {
	return urfave.Command(param.LogE, param.LDFlags, &cli.Command{ //nolint:wrapcheck
		Name:           "aqua",
		Usage:          "Version Manager of CLI. https://aquaproj.github.io/",
		ExitErrHandler: exitErrHandlerFunc,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level",
				Sources: cli.EnvVars("AQUA_LOG_LEVEL"),
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "configuration file path",
				Sources: cli.EnvVars("AQUA_CONFIG"),
			},
			&cli.BoolFlag{
				Name:    "disable-cosign",
				Usage:   "Disable Cosign verification",
				Sources: cli.EnvVars("AQUA_DISABLE_COSIGN"),
			},
			&cli.BoolFlag{
				Name:    "disable-slsa",
				Usage:   "Disable SLSA verification",
				Sources: cli.EnvVars("AQUA_DISABLE_SLSA"),
			},
			&cli.BoolFlag{
				Name:    "disable-github-artifact-attestation",
				Usage:   "Disable GitHub Artifact Attestations verification",
				Sources: cli.EnvVars("AQUA_DISABLE_GITHUB_ARTIFACT_ATTESTATION"),
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
		Commands: commands(
			param,
			initcmd.New,
			install.New,
			generate.New,
			updateaqua.New,
			upc.New,
			update.New,
			which.New,
			info.New,
			remove.New,
			vacuum.New,
			cp.New,
			cpolicy.New,
			cpolicy.NewInitPolicy,
			exec.New,
			list.New,
			genr.New,
			root.New,
		),
	}).Run(ctx, args)
}

func exitErrHandlerFunc(_ context.Context, cmd *cli.Command, err error) {
	if cmd.Name != "exec" {
		cli.HandleExitCoder(err)
		return
	}
}
