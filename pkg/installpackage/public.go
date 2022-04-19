package installpackage

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/aquaproj/aqua/pkg/runtime"
)

type Installer interface {
	InstallPackage(ctx context.Context, pkgInfo *config.PackageInfo, pkg *config.Package, isTest bool) error
	InstallPackages(ctx context.Context, cfg *config.Config, registries map[string]*config.RegistryContent, binDir string, onlyLink, isTest bool) error
	InstallProxy(ctx context.Context) error
}

func New(rootDir config.RootDir, logger *log.Logger, downloader download.PackageDownloader, rt *runtime.Runtime) Installer {
	return &installer{
		logger:            logger,
		rootDir:           string(rootDir),
		packageDownloader: downloader,
		runtime:           rt,
	}
}
