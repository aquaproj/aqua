//go:build wireinject
// +build wireinject

package controller

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	cexec "github.com/aquaproj/aqua/pkg/controller/exec"
	"github.com/aquaproj/aqua/pkg/controller/generate"
	"github.com/aquaproj/aqua/pkg/controller/initcmd"
	"github.com/aquaproj/aqua/pkg/controller/install"
	"github.com/aquaproj/aqua/pkg/controller/list"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/aquaproj/aqua/pkg/github"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/wire"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func InitializeListCommandController(ctx context.Context, param *config.Param) *list.Controller {
	wire.Build(list.NewController, finder.NewConfigFinder, github.New, registry.New, download.NewRegistryDownloader, reader.New, afero.NewOsFs)
	return &list.Controller{}
}

func InitializeInitCommandController(ctx context.Context, param *config.Param) *initcmd.Controller {
	wire.Build(initcmd.New, github.New, afero.NewOsFs)
	return &initcmd.Controller{}
}

func InitializeGenerateCommandController(ctx context.Context, param *config.Param) *generate.Controller {
	wire.Build(generate.New, finder.NewConfigFinder, github.New, registry.New, download.NewRegistryDownloader, reader.New, afero.NewOsFs, generate.NewFuzzyFinder)
	return &generate.Controller{}
}

func InitializeInstallCommandController(ctx context.Context, param *config.Param) *install.Controller {
	wire.Build(install.New, finder.NewConfigFinder, github.New, registry.New, download.NewRegistryDownloader, reader.New, installpackage.New, download.NewPackageDownloader, runtime.New, afero.NewOsFs, link.New)
	return &install.Controller{}
}

func InitializeWhichCommandController(ctx context.Context, param *config.Param) which.Controller {
	wire.Build(which.New, finder.NewConfigFinder, github.New, registry.New, download.NewRegistryDownloader, reader.New, runtime.New, osenv.New, afero.NewOsFs)
	return nil
}

func InitializeExecCommandController(ctx context.Context, param *config.Param) *cexec.Controller {
	wire.Build(cexec.New, finder.NewConfigFinder, download.NewPackageDownloader, installpackage.New, github.New, registry.New, download.NewRegistryDownloader, reader.New, which.New, runtime.New, exec.New, osenv.New, afero.NewOsFs, link.New)
	return &cexec.Controller{}
}
