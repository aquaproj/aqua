package installpackage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type CargoPackageInstaller interface {
	Install(ctx context.Context, logger *slog.Logger, crate, version, root string, opts *registry.Cargo) error
}

type CargoPackageInstallerImpl struct {
	exec    Executor
	cleaner Cleaner
}

type Cleaner interface {
	RemoveAll(name string) (err error)
}

func NewCargoPackageInstallerImpl(exec Executor, cleaner Cleaner) *CargoPackageInstallerImpl {
	return &CargoPackageInstallerImpl{
		exec:    exec,
		cleaner: cleaner,
	}
}

func getCargoArgs(version string, opts *registry.Cargo) []string {
	args := []string{"install", "--version", version}
	if opts != nil {
		if opts.AllFeatures {
			args = append(args, "--all-features")
		} else if len(opts.Features) != 0 {
			args = append(args, "--features", strings.Join(opts.Features, ","))
		}
		if opts.Locked {
			args = append(args, "--locked")
		}
	}
	return args
}

func (is *CargoPackageInstallerImpl) Install(ctx context.Context, logger *slog.Logger, crate, version, root string, opts *registry.Cargo) error {
	args := getCargoArgs(version, opts)
	if _, err := is.exec.ExecStderr(osexec.Command(ctx, "cargo", append(args, "--root", root, crate)...)); err != nil {
		// Clean up root
		logger := logger.With("install_dir", root)
		logger.Info("removing the install directory because the installation failed")
		if err := is.cleaner.RemoveAll(root); err != nil {
			slogerr.WithError(logger, err).Error("aqua tried to remove the install directory because the installation failed, but it failed")
		}
		return fmt.Errorf("install a crate: %w", slogerr.With(err,
			"doc", "https://aquaproj.github.io/docs/reference/codes/005",
			"cargo_command", strings.Join(append(append([]string{"cargo"}, args...), crate), " ")))
	}
	return nil
}

func (is *Installer) downloadCargo(ctx context.Context, logger *slog.Logger, pkg *config.Package, root string) error {
	cargoOpts := pkg.PackageInfo.Cargo
	if cargoOpts != nil {
		if cargoOpts.AllFeatures {
			logger = logger.With("cargo_all_features", true)
		} else if len(cargoOpts.Features) != 0 {
			logger = logger.With("cargo_features", strings.Join(cargoOpts.Features, ","))
		}
	}
	logger.Info("Installing a crate")
	crate := pkg.PackageInfo.Crate
	version := strings.TrimPrefix(strings.TrimPrefix(pkg.Package.Version, pkg.PackageInfo.VersionPrefix), "v")
	if err := is.cargoPackageInstaller.Install(ctx, logger, crate, version, root, cargoOpts); err != nil {
		return fmt.Errorf("cargo install: %w", err)
	}
	return nil
}
