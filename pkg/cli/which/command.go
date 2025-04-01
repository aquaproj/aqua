package which

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/urfave/cli/v3"
)

type command struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:      "which",
		Usage:     "Output the absolute file path of the given command",
		ArgsUsage: `<command name>`,
		Description: `Output the absolute file path of the given command
e.g.
$ aqua which gh
/home/foo/.aqua/pkgs/github_release/github.com/cli/cli/v2.4.0/gh_2.4.0_macOS_amd64.tar.gz/gh_2.4.0_macOS_amd64/bin/gh

If the command isn't found in the configuration files, aqua searches the command in the environment variable PATH

$ aqua which ls
/bin/ls

If the command isn't found, exits with non zero exit code.

$ aqua which foo
FATA[0000] aqua failed                                   aqua_version=0.8.6 error="command is not found" exe_name=foo program=aqua

If you want the package version, "--version" option is useful.

$ aqua which --version gh
v2.4.0
`,
		Action: i.action,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Output the given package version",
			},
		},
	}
}

func (i *command) action(c *cli.Context) error {
	profiler, err := profile.Start(c)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(c, i.r.LogE, "which", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeWhichCommandController(c.Context, param, http.DefaultClient, i.r.Runtime)
	exeName, _, err := ParseExecArgs(c.Args().Slice())
	if err != nil {
		return err
	}
	logE := i.r.LogE.WithField("exe_name", exeName)
	which, err := ctrl.Which(c.Context, logE, param, exeName)
	if err != nil {
		return logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
			"exe_name": exeName,
		})
	}
	if !param.ShowVersion {
		fmt.Fprintln(os.Stdout, which.ExePath)
		return nil
	}
	if which.Package == nil {
		return logerr.WithFields(errors.New("aqua can't get the command version because the command isn't managed by aqua"), logrus.Fields{ //nolint:wrapcheck
			"exe_name": exeName,
		})
	}
	fmt.Fprintln(os.Stdout, which.Package.Package.Version)
	return nil
}

var errCommandIsRequired = errors.New("command is required")

func ParseExecArgs(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, errCommandIsRequired
	}
	return filepath.Base(args[0]), args[1:], nil
}
