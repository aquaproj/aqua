package main

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/cli"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	_ "golang.org/x/crypto/x509roots/fallback"
)

var version = ""

func main() {
	urfave.Main("aqua", version, cli.Run, "env", runtime.New(context.Background()).Env())
}
