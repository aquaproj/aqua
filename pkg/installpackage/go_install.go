package installpackage

import (
	"context"
	"fmt"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/sirupsen/logrus"
)

type GoInstallInstaller interface {
	Install(ctx context.Context, path, gobin string) error
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
	cmd := osexec.Command(ctx, "go", "install", path)
	cmd.Env = append(os.Environ(), "GOBIN="+gobin)
	_, err := is.exec.Exec(cmd, nil)
	if err != nil {
		return fmt.Errorf("install a go package: %w", err)
	}
	return nil
}

func (is *Installer) downloadGoInstall(ctx context.Context, logE *logrus.Entry, pkg *config.Package, dest string) error {
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
