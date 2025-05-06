package token

import (
	"context"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/controller/rmtoken"
	"github.com/aquaproj/aqua/v2/pkg/controller/settoken"
	"github.com/aquaproj/aqua/v2/pkg/keyring"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func New(logE *logrus.Entry) *cli.Command {
	r := &runner{
		logE: logE,
	}
	return r.Command()
}

type runner struct {
	logE *logrus.Entry
}

func (r *runner) Command() *cli.Command {
	return &cli.Command{
		Name:        "token",
		Usage:       "Manage GitHub Access token",
		Description: `Manage GitHub Access token by keyring.`,
		Commands: []*cli.Command{
			{
				Name:        "set",
				Usage:       "Set GitHub Access token",
				Description: `Set GitHub Access token to keyring.`,
				Action:      r.action,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "stdin",
						Usage: "Read GitHub Access token from stdin",
					},
				},
			},
			{
				Name:        "remove",
				Aliases:     []string{"rm"},
				Usage:       "Remove GitHub Access token",
				Description: `Remove GitHub Access token from keyring.`,
				Action:      r.remove,
			},
		},
	}
}

func (r *runner) action(_ context.Context, c *cli.Command) error {
	term := settoken.NewPasswordReader(os.Stdout)
	tokenManager := keyring.NewTokenManager()
	ctrl := settoken.New(&settoken.Param{
		IsStdin: c.Bool("stdin"),
		Stdin:   os.Stdin,
	}, term, tokenManager)
	return ctrl.Set(r.logE) //nolint:wrapcheck
}

func (r *runner) remove(_ context.Context, _ *cli.Command) error {
	tokenManager := keyring.NewTokenManager()
	ctrl := rmtoken.New(&rmtoken.Param{}, tokenManager)
	return ctrl.Remove(r.logE) //nolint:wrapcheck
}
