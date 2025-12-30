package main

import (
	"github.com/aquaproj/aqua/v2/pkg/cli"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
)

var version = ""

type HasExitCode interface {
	ExitCode() int
}

func main() {
	urfave.Main("aqua", version, cli.Run)
}
