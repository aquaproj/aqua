package installpackage

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type Installer interface {
	InstallPackage(ctx context.Context, pkgInfo *config.PackageInfo, pkg *config.Package, isTest bool, logE *logrus.Entry) error
	InstallPackages(ctx context.Context, cfg *config.Config, registries map[string]*config.RegistryContent, binDir string, onlyLink, isTest bool, logE *logrus.Entry) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
}

func New(rootDir config.RootDir, downloader download.PackageDownloader, rt *runtime.Runtime) Installer {
	return &installer{
		rootDir:           string(rootDir),
		packageDownloader: downloader,
		runtime:           rt,
	}
}
