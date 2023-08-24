package cp

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/sirupsen/logrus"
)

type PackageInstaller interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, param *installpackage.ParamInstallPackage) error
	InstallPackages(ctx context.Context, logE *logrus.Entry, param *installpackage.ParamInstallPackages) error
	SetCopyDir(copyDir string)
	Copy(dest, src string) error
	WaitExe(ctx context.Context, logE *logrus.Entry, exePath string) error
}

type MockPackageInstaller struct{}

func (is *MockPackageInstaller) InstallPackage(ctx context.Context, logE *logrus.Entry, param *installpackage.ParamInstallPackage) error {
	return nil
}

func (is *MockPackageInstaller) InstallPackages(ctx context.Context, logE *logrus.Entry, param *installpackage.ParamInstallPackages) error {
	return nil
}

func (is *MockPackageInstaller) SetCopyDir(copyDir string) {
}

func (is *MockPackageInstaller) Copy(dest, src string) error {
	return nil
}

func (is *MockPackageInstaller) WaitExe(ctx context.Context, logE *logrus.Entry, exePath string) error {
	return nil
}
