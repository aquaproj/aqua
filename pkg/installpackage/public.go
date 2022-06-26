package installpackage

import (
	"context"

	"github.com/clivm/clivm/pkg/config"
	"github.com/clivm/clivm/pkg/config/clivm"
	"github.com/clivm/clivm/pkg/config/registry"
	"github.com/clivm/clivm/pkg/download"
	"github.com/clivm/clivm/pkg/exec"
	"github.com/clivm/clivm/pkg/link"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Installer interface {
	InstallPackage(ctx context.Context, pkg *config.Package, isTest bool, logE *logrus.Entry) error
	InstallPackages(ctx context.Context, cfg *clivm.Config, registries map[string]*registry.Config, binDir string, onlyLink, isTest bool, logE *logrus.Entry) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
}

func New(param *config.Param, downloader download.PackageDownloader, rt *runtime.Runtime, fs afero.Fs, linker link.Linker, executor exec.Executor) Installer {
	return &installer{
		rootDir:           param.RootDir,
		maxParallelism:    param.MaxParallelism,
		packageDownloader: downloader,
		runtime:           rt,
		fs:                fs,
		linker:            linker,
		executor:          executor,
	}
}
