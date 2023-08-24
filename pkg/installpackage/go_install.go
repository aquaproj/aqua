package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

type GoInstallInstaller interface {
	Install(ctx context.Context, path, gobin string) error
}

type MockGoInstallInstaller struct {
	Err error
}

func (mock *MockGoInstallInstaller) Install(ctx context.Context, path, gobin string) error {
	return mock.Err
}

type GoInstallInstallerImpl struct {
	exec Executor
}

func NewGoInstallInstallerImpl(exec Executor) *GoInstallInstallerImpl {
	return &GoInstallInstallerImpl{
		exec: exec,
	}
}

func (is *GoInstallInstallerImpl) Install(ctx context.Context, path, gobin string) error {
	_, err := is.exec.ExecWithEnvs(ctx, "go", []string{"install", path}, []string{fmt.Sprintf("GOBIN=%s", gobin)})
	if err != nil {
		return fmt.Errorf("install a go package: %w", err)
	}
	return nil
}

func (is *InstallerImpl) downloadGoInstall(ctx context.Context, pkg *config.Package, dest string, logE *logrus.Entry) error {
	p, err := pkg.RenderPath()
	if err != nil {
		return fmt.Errorf("render Go Module Path: %w", err)
	}
	goPkgPath := p + "@" + pkg.Package.Version
	logE.WithFields(logrus.Fields{
		"gobin":           dest,
		"go_package_path": goPkgPath,
	}).Info("Installing a Go tool")
	if err := is.goInstallInstaller.Install(ctx, goPkgPath, dest); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}
