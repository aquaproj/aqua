package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/aquaproj/aqua/v2/pkg/cli"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/log"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

type HasExitCode interface {
	ExitCode() int
}

func main() {
	rt := runtime.New()

	logE := log.New("aqua", version).WithField("env", rt.Env())
	err := core(logE, rt)
	if err != nil {
		var hasExitCode HasExitCode
		if errors.As(err, &hasExitCode) {
			code := hasExitCode.ExitCode()
			logerr.WithError(logE.WithField("exit_code", code), err).Debug("command failed")
			os.Exit(code)
		}

		logerr.WithError(logE, err).Fatal("aqua failed")
	}
}

func core(logE *logrus.Entry, rt *runtime.Runtime) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return cli.Run(ctx, &util.Param{ //nolint:wrapcheck
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		LDFlags: &urfave.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
		LogE:    logE,
		Runtime: rt,
	}, os.Args...)
}
