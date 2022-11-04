package cp

import (
	"context"

	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/sirupsen/logrus"
)

type PackageInstaller interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, param *domain.ParamInstallPackage) error
	InstallPackages(ctx context.Context, logE *logrus.Entry, param *domain.ParamInstallPackages) error
	SetCopyDir(copyDir string)
	Copy(dest, src string) error
	WaitExe(ctx context.Context, logE *logrus.Entry, exePath string) error
}

type MockPackageInstaller struct{}

func (inst *MockPackageInstaller) InstallPackage(ctx context.Context, logE *logrus.Entry, param *domain.ParamInstallPackage) error {
	return nil
}

func (inst *MockPackageInstaller) InstallPackages(ctx context.Context, logE *logrus.Entry, param *domain.ParamInstallPackages) error {
	return nil
}

func (inst *MockPackageInstaller) SetCopyDir(copyDir string) {
}

func (inst *MockPackageInstaller) Copy(dest, src string) error {
	return nil
}

func (inst *MockPackageInstaller) WaitExe(ctx context.Context, logE *logrus.Entry, exePath string) error {
	return nil
}
