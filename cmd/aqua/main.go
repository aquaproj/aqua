package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/aqua/pkg/cli"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	if err := core(); err != nil {
		logrus.Fatal(err)
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return runner.Run(ctx, os.Args...) //nolint:wrapcheck
}
