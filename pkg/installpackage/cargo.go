package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
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

func (inst *CargoPackageInstallerImpl) Install(ctx context.Context, logE *logrus.Entry, crate, version, root string, opts *registry.Cargo) error {
	args := []string{"install", "--version", version, "--root", root}
	if opts != nil {
		if opts.AllFeatures {
			args = append(args, "--all-features")
		} else if len(opts.Features) != 0 {
			args = append(append(args, "--features"), opts.Features...)
		}
	}
	_, err := inst.exec.Exec(ctx, "cargo", append(args, crate)...)
	if err != nil {
		// Clean up root
		logE := logE.WithField("install_dir", root)
		logE.Info("removing the install directory because the installation failed")
		if err := inst.cleaner.RemoveAll(root); err != nil {
			logE.WithError(err).Error("aqua tried to remove the install directory because the installation failed, but it failed")
		}
		return fmt.Errorf("install a crate: %w", err)
	}
	return nil
}
