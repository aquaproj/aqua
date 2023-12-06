package cp

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/sirupsen/logrus"
)

type MockInstaller struct {
	Err error
}

func (is *MockInstaller) Install(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	return is.Err
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
