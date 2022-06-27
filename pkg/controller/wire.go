//go:build wireinject
// +build wireinject

package controller

import (
	"context"
	"net/http"

	"github.com/clivm/clivm/pkg/config"
	finder "github.com/clivm/clivm/pkg/config-finder"
	reader "github.com/clivm/clivm/pkg/config-reader"
	cexec "github.com/clivm/clivm/pkg/controller/exec"
	"github.com/clivm/clivm/pkg/controller/generate"
	genrgst "github.com/clivm/clivm/pkg/controller/generate-registry"
	"github.com/clivm/clivm/pkg/controller/initcmd"
	"github.com/clivm/clivm/pkg/controller/install"
	"github.com/clivm/clivm/pkg/controller/list"
	"github.com/clivm/clivm/pkg/controller/which"
	"github.com/clivm/clivm/pkg/download"
	"github.com/clivm/clivm/pkg/exec"
	"github.com/clivm/clivm/pkg/github"
	registry "github.com/clivm/clivm/pkg/install-registry"
	"github.com/clivm/clivm/pkg/installpackage"
	"github.com/clivm/clivm/pkg/link"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/google/wire"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func InitializeListCommandController(ctx context.Context, param *config.Param, httpClient *http.Client) *list.Controller {
	wire.Build(list.NewController, finder.NewConfigFinder, github.New, registry.New, download.NewRegistryDownloader, reader.New, afero.NewOsFs, download.NewHTTPDownloader)
	return &list.Controller{}
}

func InitializeGenerateRegistryCommandController(ctx context.Context, param *config.Param, httpClient *http.Client) *genrgst.Controller {
	wire.Build(genrgst.NewController, github.New, afero.NewOsFs)
	return &genrgst.Controller{}
}

func InitializeInitCommandController(ctx context.Context, param *config.Param) *initcmd.Controller {
	wire.Build(initcmd.New, github.New, afero.NewOsFs)
	return &initcmd.Controller{}
}

func InitializeGenerateCommandController(ctx context.Context, param *config.Param, httpClient *http.Client) *generate.Controller {
	wire.Build(generate.New, finder.NewConfigFinder, github.New, registry.New, download.NewRegistryDownloader, reader.New, afero.NewOsFs, generate.NewFuzzyFinder, download.NewHTTPDownloader)
	return &generate.Controller{}
}

func InitializeInstallCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *install.Controller {
	wire.Build(install.New, finder.NewConfigFinder, github.New, registry.New, download.NewRegistryDownloader, reader.New, installpackage.New, download.NewPackageDownloader, afero.NewOsFs, link.New, download.NewHTTPDownloader, exec.New)
	return &install.Controller{}
}

func InitializeWhichCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) which.Controller {
	wire.Build(which.New, finder.NewConfigFinder, github.New, registry.New, download.NewRegistryDownloader, reader.New, osenv.New, afero.NewOsFs, download.NewHTTPDownloader, link.New)
	return nil
}

func InitializeExecCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *cexec.Controller {
	wire.Build(cexec.New, finder.NewConfigFinder, download.NewPackageDownloader, installpackage.New, github.New, registry.New, download.NewRegistryDownloader, reader.New, which.New, exec.New, osenv.New, afero.NewOsFs, link.New, download.NewHTTPDownloader)
	return &cexec.Controller{}
}
