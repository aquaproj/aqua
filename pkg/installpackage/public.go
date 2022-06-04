package installpackage

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/pkgtype"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Installer interface {
	InstallPackage(ctx context.Context, pkgInfo *config.PackageInfo, pkg *config.Package, isTest bool, logE *logrus.Entry) error
	InstallPackages(ctx context.Context, cfg *config.Config, registries map[string]*config.RegistryContent, binDir string, onlyLink, isTest bool, logE *logrus.Entry) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
}

func New(param *config.Param, rt *runtime.Runtime, fs afero.Fs, linker link.Linker, executor exec.Executor, pkgTypes pkgtype.Packages) Installer {
	return &installer{
		rootDir:        param.RootDir,
		maxParallelism: param.MaxParallelism,
		runtime:        rt,
		fs:             fs,
		linker:         linker,
		executor:       executor,
		installers:     pkgTypes,
	}
}
