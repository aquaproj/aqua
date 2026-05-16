package main

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/cli"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
)

var version = ""

func main() {
	urfave.Main("aqua", version, cli.Run, "env", runtime.New(context.Background()).Env())
}
