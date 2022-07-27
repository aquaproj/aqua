//go:build wireinject
// +build wireinject

package controller

import (
	"context"
	"net/http"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	cexec "github.com/aquaproj/aqua/pkg/controller/exec"
	"github.com/aquaproj/aqua/pkg/controller/generate"
	genrgst "github.com/aquaproj/aqua/pkg/controller/generate-registry"
	"github.com/aquaproj/aqua/pkg/controller/initcmd"
	"github.com/aquaproj/aqua/pkg/controller/install"
	"github.com/aquaproj/aqua/pkg/controller/list"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/domain"
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

func InitializeListCommandController(ctx context.Context, param *config.Param, httpClient *http.Client) *list.Controller {
	wire.Build(
		list.NewController,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(list.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.RepositoriesService), new(*github.RepositoriesService)),
		),
		registry.New,
		wire.NewSet(
			download.NewRegistryDownloader,
			wire.Bind(new(domain.RegistryDownloader), new(*download.RegistryDownloader)),
		),
		reader.New,
		afero.NewOsFs,
		download.NewHTTPDownloader,
	)
	return &list.Controller{}
}

func InitializeGenerateRegistryCommandController(ctx context.Context, param *config.Param, httpClient *http.Client) *genrgst.Controller {
	wire.Build(
		genrgst.NewController,
		wire.NewSet(
			github.New,
			wire.Bind(new(genrgst.RepositoriesService), new(*github.RepositoriesService)),
		),
		afero.NewOsFs,
	)
	return &genrgst.Controller{}
}

func InitializeInitCommandController(ctx context.Context, param *config.Param) *initcmd.Controller {
	wire.Build(
		initcmd.New,
		wire.NewSet(
			github.New,
			wire.Bind(new(initcmd.RepositoriesService), new(*github.RepositoriesService)),
		),
		afero.NewOsFs,
	)
	return &initcmd.Controller{}
}

func InitializeGenerateCommandController(ctx context.Context, param *config.Param, httpClient *http.Client) *generate.Controller {
	wire.Build(
		generate.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(generate.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(generate.RepositoriesService), new(*github.RepositoriesService)),
			wire.Bind(new(download.RepositoriesService), new(*github.RepositoriesService)),
		),
		registry.New,
		wire.NewSet(
			download.NewRegistryDownloader,
			wire.Bind(new(domain.RegistryDownloader), new(*download.RegistryDownloader)),
		),
		reader.New,
		afero.NewOsFs,
		generate.NewFuzzyFinder,
		generate.NewVersionSelector,
		download.NewHTTPDownloader,
	)
	return &generate.Controller{}
}

func InitializeInstallCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *install.Controller {
	wire.Build(
		install.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(install.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.RepositoriesService), new(*github.RepositoriesService)),
		),
		registry.New,
		wire.NewSet(
			download.NewRegistryDownloader,
			wire.Bind(new(domain.RegistryDownloader), new(*download.RegistryDownloader)),
		),
		reader.New,
		installpackage.New,
		wire.NewSet(
			download.NewPackageDownloader,
			wire.Bind(new(domain.PackageDownloader), new(*download.PackageDownloader)),
		),
		afero.NewOsFs,
		link.New,
		download.NewHTTPDownloader,
		wire.NewSet(
			exec.New,
			wire.Bind(new(installpackage.Executor), new(*exec.Executor)),
		),
	)
	return &install.Controller{}
}

func InitializeWhichCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) which.Controller {
	wire.Build(
		which.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.RepositoriesService), new(*github.RepositoriesService)),
		),
		registry.New,
		wire.NewSet(
			download.NewRegistryDownloader,
			wire.Bind(new(domain.RegistryDownloader), new(*download.RegistryDownloader)),
		),
		reader.New,
		osenv.New,
		afero.NewOsFs,
		download.NewHTTPDownloader,
		link.New,
	)
	return nil
}

func InitializeExecCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *cexec.Controller {
	wire.Build(
		cexec.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			download.NewPackageDownloader,
			wire.Bind(new(domain.PackageDownloader), new(*download.PackageDownloader)),
		),
		installpackage.New,
		wire.NewSet(
			github.New,
			wire.Bind(new(download.RepositoriesService), new(*github.RepositoriesService)),
		),
		registry.New,
		wire.NewSet(
			download.NewRegistryDownloader,
			wire.Bind(new(domain.RegistryDownloader), new(*download.RegistryDownloader)),
		),
		reader.New,
		which.New,
		wire.NewSet(
			exec.New,
			wire.Bind(new(installpackage.Executor), new(*exec.Executor)),
			wire.Bind(new(cexec.Executor), new(*exec.Executor)),
		),
		osenv.New,
		afero.NewOsFs,
		link.New,
		download.NewHTTPDownloader,
	)
	return &cexec.Controller{}
}
