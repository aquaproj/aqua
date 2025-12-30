package cp

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
)

type MockInstaller struct {
	Err error
}

func (is *MockInstaller) Install(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	return is.Err
}

type MockPackageInstaller struct{}

func (is *MockPackageInstaller) InstallPackage(ctx context.Context, logger *slog.Logger, param *installpackage.ParamInstallPackage) error {
	return nil
}

func (is *MockPackageInstaller) InstallPackages(ctx context.Context, logger *slog.Logger, param *installpackage.ParamInstallPackages) error {
	return nil
}

func (is *MockPackageInstaller) SetCopyDir(copyDir string) {
}

func (is *MockPackageInstaller) Copy(dest, src string) error {
	return nil
}

func (is *MockPackageInstaller) WaitExe(ctx context.Context, logger *slog.Logger, exePath string) error {
	return nil
}
