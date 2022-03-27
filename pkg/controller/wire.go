//go:build wireinject
// +build wireinject

package controller

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/github"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/google/wire"
)

func NewController(ctx context.Context, aquaVersion string, param *config.Param) (*Controller, error) {
	wire.Build(New, finder.NewConfigFinder, log.NewLogger, download.NewPackageDownloader, installpackage.New, github.NewGitHub, config.NewRootDir, registry.New, download.NewRegistryDownloader)
	return &Controller{}, nil
}
