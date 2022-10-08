package cp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const dirPermission os.FileMode = 0o775

type Controller struct {
	packageInstaller PackageInstaller
	rootDir          string
	fs               afero.Fs
	runtime          *runtime.Runtime
	which            domain.WhichController
	installer        Installer
}

type PackageInstaller interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, param *domain.ParamInstallPackage) error
	InstallPackages(ctx context.Context, logE *logrus.Entry, param *domain.ParamInstallPackages) error
	SetCopyDir(copyDir string)
	Copy(dest, src string) error
}

type Installer interface {
	Install(ctx context.Context, logE *logrus.Entry, param *config.Param) error
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

func New(param *config.Param, pkgInstaller PackageInstaller, fs afero.Fs, rt *runtime.Runtime, whichCtrl domain.WhichController, installer Installer) *Controller {
	return &Controller{
		rootDir:          param.RootDir,
		packageInstaller: pkgInstaller,
		fs:               fs,
		runtime:          rt,
		which:            whichCtrl,
		installer:        installer,
	}
}

var errCopyFailure = errors.New("it failed to copy some tools")

func (ctrl *Controller) Copy(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if err := ctrl.fs.MkdirAll(param.Dest, dirPermission); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}
	if len(param.Args) == 0 {
		return ctrl.installer.Install(ctx, logE, param) //nolint:wrapcheck
	}

	maxInstallChan := make(chan struct{}, param.MaxParallelism)
	var wg sync.WaitGroup
	wg.Add(len(param.Args))
	var flagMutex sync.Mutex
	failed := false
	handleFailure := func() {
		flagMutex.Lock()
		failed = true
		flagMutex.Unlock()
	}

	ctrl.packageInstaller.SetCopyDir("")

	for _, exeName := range param.Args {
		go func(exeName string) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			defer func() {
				<-maxInstallChan
			}()
			logE := logE.WithField("exe_name", exeName)
			if err := ctrl.installAndCopy(ctx, logE, param, exeName); err != nil {
				logerr.WithError(logE, err).Error("install the package")
				handleFailure()
				return
			}
		}(exeName)
	}
	wg.Wait()
	if failed {
		return errCopyFailure
	}
	return nil
}

func (ctrl *Controller) installAndCopy(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string) error {
	findResult, err := ctrl.which.Which(ctx, param, exeName, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}
	logE = logE.WithField("exe_path", findResult.ExePath)
	if findResult.Package != nil {
		logE = logE.WithField("package", findResult.Package.Package.Name)
		if err := ctrl.install(ctx, logE, findResult); err != nil {
			return err
		}
	}

	if err := ctrl.copy(logE, param, findResult, exeName); err != nil {
		return err
	}
	return nil
}
