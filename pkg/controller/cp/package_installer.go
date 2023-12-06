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

func (is *MockPackageInstaller) InstallPackage(context.Context, *logrus.Entry, *installpackage.ParamInstallPackage) error {
	return nil
}

func (is *MockPackageInstaller) InstallPackages(context.Context, *logrus.Entry, *installpackage.ParamInstallPackages) error {
	return nil
}

func (is *MockPackageInstaller) SetCopyDir(string) {
}

func (is *MockPackageInstaller) Copy(string, string) error {
	return nil
}

func (is *MockPackageInstaller) WaitExe(context.Context, *logrus.Entry, string) error {
	return nil
}
