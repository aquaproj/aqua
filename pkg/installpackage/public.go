package installpackage

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Installer interface {
	InstallPackage(ctx context.Context, pkg *config.Package, logE *logrus.Entry) error
	InstallPackages(ctx context.Context, cfg *aqua.Config, registries map[string]*registry.Config, logE *logrus.Entry) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
}

type Executor interface {
	GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error)
	GoInstall(ctx context.Context, path, gobin string) (int, error)
}

func New(param *config.Param, downloader domain.PackageDownloader, rt *runtime.Runtime, fs afero.Fs, linker link.Linker, executor Executor) Installer {
	return &installer{
		rootDir:           param.RootDir,
		maxParallelism:    param.MaxParallelism,
		packageDownloader: downloader,
		runtime:           rt,
		fs:                fs,
		linker:            linker,
		executor:          executor,
		progressBar:       param.ProgressBar,
		isTest:            param.IsTest,
		onlyLink:          param.OnlyLink,
	}
}
