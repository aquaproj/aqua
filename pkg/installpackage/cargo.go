package installpackage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type CargoPackageInstaller interface {
	Install(ctx context.Context, logE *logrus.Entry, crate, version, root string, opts *registry.Cargo) error
}

type MockCargoPackageInstaller struct {
	Err error
}

func (mock *MockCargoPackageInstaller) Install(ctx context.Context, logE *logrus.Entry, crate, version, root string, opts *registry.Cargo) error {
	return mock.Err
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
	}
	return args
}

func (inst *CargoPackageInstallerImpl) Install(ctx context.Context, logE *logrus.Entry, crate, version, root string, opts *registry.Cargo) error {
	args := getCargoArgs(version, opts)
	if _, err := inst.exec.Exec(ctx, "cargo", append(args, "--root", root, crate)...); err != nil {
		// Clean up root
		logE := logE.WithField("install_dir", root)
		logE.Info("removing the install directory because the installation failed")
		if err := inst.cleaner.RemoveAll(root); err != nil {
			logE.WithError(err).Error("aqua tried to remove the install directory because the installation failed, but it failed")
		}
		return fmt.Errorf("install a crate: %w", logerr.WithFields(err, logrus.Fields{
			"doc":           "https://aquaproj.github.io/docs/reference/codes/005",
			"cargo_command": strings.Join(append(append([]string{"cargo"}, args...), crate), " "),
		}))
	}
	return nil
}

func (inst *InstallerImpl) downloadCargo(ctx context.Context, logE *logrus.Entry, pkg *config.Package, root string) error {
	cargoOpts := pkg.PackageInfo.Cargo
	if cargoOpts != nil {
		if cargoOpts.AllFeatures {
			logE = logE.WithField("cargo_all_features", true)
		} else if len(cargoOpts.Features) != 0 {
			logE = logE.WithField("cargo_features", strings.Join(cargoOpts.Features, ","))
		}
	}
	logE.Info("Installing a crate")
	crate := *pkg.PackageInfo.Crate
	version := pkg.Package.Version
	if err := inst.cargoPackageInstaller.Install(ctx, logE, crate, version, root, cargoOpts); err != nil {
		return fmt.Errorf("cargo install: %w", err)
	}
	return nil
}
