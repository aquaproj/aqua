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
	"github.com/aquaproj/aqua/pkg/controller/initcmd"
	"github.com/aquaproj/aqua/pkg/controller/install"
	"github.com/aquaproj/aqua/pkg/controller/list"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/github/archive"
	"github.com/aquaproj/aqua/pkg/github/content"
	"github.com/aquaproj/aqua/pkg/github/release"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/pkgtype"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubarchive"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubcontent"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubrelease"
	"github.com/aquaproj/aqua/pkg/pkgtype/gobuild"
	"github.com/aquaproj/aqua/pkg/pkgtype/goinstall"
	httpinstall "github.com/aquaproj/aqua/pkg/pkgtype/http"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/wire"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func InitializeListCommandController(ctx context.Context, param *config.Param, httpClient *http.Client) *list.Controller {
	wire.Build(
		list.NewController,
		finder.NewConfigFinder,
		github.New,
		registry.New,
		download.NewRegistryDownloader,
		reader.New,
		afero.NewOsFs,
		download.NewHTTPDownloader,
	)
	return &list.Controller{}
}

func InitializeInitCommandController(ctx context.Context, param *config.Param) *initcmd.Controller {
	wire.Build(
		initcmd.New,
		github.New,
		afero.NewOsFs,
	)
	return &initcmd.Controller{}
}

func InitializeGenerateCommandController(ctx context.Context, param *config.Param, httpClient *http.Client) *generate.Controller {
	wire.Build(
		generate.New,
		finder.NewConfigFinder,
		github.New,
		registry.New,
		download.NewRegistryDownloader,
		reader.New,
		afero.NewOsFs,
		generate.NewFuzzyFinder,
		download.NewHTTPDownloader,
	)
	return &generate.Controller{}
}

func InitializeInstallCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *install.Controller {
	wire.Build(
		install.New,
		finder.NewConfigFinder,
		github.New, github.NewArchiveClient,
		archive.New, content.New, release.New,
		gobuild.New, goinstall.New, httpinstall.New, githubrelease.New, githubcontent.New, githubarchive.New,
		gobuild.NewGoBuilder, goinstall.NewGoInstaller,
		registry.New,
		download.NewRegistryDownloader,
		reader.New,
		installpackage.New,
		afero.NewOsFs,
		link.New,
		download.NewHTTPDownloader,
		exec.New,
		pkgtype.New,
	)
	return &install.Controller{}
}

func InitializeWhichCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) which.Controller {
	wire.Build(which.New,
		finder.NewConfigFinder,
		github.New, github.NewArchiveClient,
		archive.New, content.New, release.New,
		gobuild.New, goinstall.New, httpinstall.New, githubrelease.New, githubcontent.New, githubarchive.New,
		gobuild.NewGoBuilder, goinstall.NewGoInstaller,
		registry.New,
		download.NewRegistryDownloader,
		reader.New,
		osenv.New,
		afero.NewOsFs,
		download.NewHTTPDownloader,
		link.New,
		pkgtype.New,
		exec.New,
	)
	return nil
}

func InitializeExecCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *cexec.Controller {
	wire.Build(cexec.New,
		finder.NewConfigFinder,
		installpackage.New,
		github.New, github.NewArchiveClient,
		archive.New, content.New, release.New,
		gobuild.New, goinstall.New, httpinstall.New, githubrelease.New, githubcontent.New, githubarchive.New,
		gobuild.NewGoBuilder, goinstall.NewGoInstaller,
		registry.New,
		download.NewRegistryDownloader,
		reader.New,
		which.New,
		exec.New,
		osenv.New,
		afero.NewOsFs,
		link.New,
		download.NewHTTPDownloader,
		pkgtype.New,
	)
	return &cexec.Controller{}
}
