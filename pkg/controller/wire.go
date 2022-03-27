//go:build wireinject
// +build wireinject

package controller

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/github"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/google/wire"
)

func NewController(ctx context.Context, aquaVersion string, param *config.Param) (*Controller, error) {
	wire.Build(New, finder.NewConfigFinder, log.NewLogger, download.NewPackageDownloader, installpackage.New, github.NewGitHub, config.NewRootDir, registry.New, download.NewRegistryDownloader, reader.New, reader.NewFileReader)
	return &Controller{}, nil
}

func InitializeListCommandController(ctx context.Context, aquaVersion string, param *config.Param) *ListCommandController {
	wire.Build(NewListCommandController, finder.NewConfigFinder, log.NewLogger, github.NewGitHub, config.NewRootDir, registry.New, download.NewRegistryDownloader, reader.New, reader.NewFileReader)
	return &ListCommandController{}
}
