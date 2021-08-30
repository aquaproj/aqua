package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/suzuki-shunsuke/aqua/pkg/cli"
	"github.com/suzuki-shunsuke/aqua/pkg/log"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
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
	if err := core(); err != nil {
		var hasExitCode HasExitCode
		if errors.As(err, &hasExitCode) {
			code := hasExitCode.ExitCode()
			logerr.WithError(log.New().WithField("exit_code", code), err).Debug("command failed")
			os.Exit(code)
		}
		logerr.WithError(log.New(), err).Fatal("aqua failed")
	}
}

func core() error {
	runner := cli.Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		LDFlags: &cli.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runner.Run(ctx, os.Args...) //nolint:wrapcheck
}
