package installpackage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
)

type DedicatedInstaller struct {
	installer *Installer
	mutex     *sync.Mutex
	pkg       func() *config.Package
	checksums *checksum.Checksums
}

func newDedicatedInstaller(installer *Installer, pkg func() *config.Package, checksums *checksum.Checksums) *DedicatedInstaller {
	return &DedicatedInstaller{
		installer: installer,
		mutex:     &sync.Mutex{},
		pkg:       pkg,
		checksums: checksums,
	}
}

func (di *DedicatedInstaller) install(ctx context.Context, logger *slog.Logger) error {
	di.mutex.Lock()
	defer di.mutex.Unlock()

	pkg := di.pkg()
	logger = logger.With(
		"package_name", pkg.Package.Name,
		"package_version", pkg.Package.Version,
	)

	pkgInfo, err := pkg.PackageInfo.Override(logger, pkg.Package.Version, di.installer.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(di.installer.runtime, di.installer.runtime.Env())
	if err != nil {
		return fmt.Errorf("check if the package is supported in the environment: %w", err)
	}
	if !supported {
		logger.Debug("the package isn't supported in the environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := di.installer.InstallPackage(ctx, logger, &ParamInstallPackage{
		Pkg:           pkg,
		Checksums:     di.checksums,
		DisablePolicy: true,
	}); err != nil {
		return err
	}

	return nil
}
