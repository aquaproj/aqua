package cp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const (
	filePermission os.FileMode = 0o755
	dirPermission  os.FileMode = 0o775
)

type Controller struct {
	packageInstaller domain.PackageInstaller
	rootDir          string
	fs               afero.Fs
	runtime          *runtime.Runtime
	which            which.Controller
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

func New(param *config.Param, pkgInstaller domain.PackageInstaller, fs afero.Fs, rt *runtime.Runtime, whichCtrl which.Controller) *Controller {
	return &Controller{
		rootDir:          param.RootDir,
		packageInstaller: pkgInstaller,
		fs:               fs,
		runtime:          rt,
		which:            whichCtrl,
	}
}

var errCopyFailure = errors.New("it failed to copy some tools")

func (ctrl *Controller) Copy(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if len(param.Args) == 0 {
		return nil
	}
	if err := ctrl.fs.MkdirAll(param.Dest, dirPermission); err != nil {
		return fmt.Errorf("create the directory: %w", err)
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

	for _, exeName := range param.Args {
		go func(exeName string) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			defer func() {
				<-maxInstallChan
			}()
			if err := ctrl.copy(ctx, logE, param, exeName); err != nil {
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

func (ctrl *Controller) copy(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string) error { //nolint:cyclop,funlen,gocognit
	which, err := ctrl.which.Which(ctx, param, exeName, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if which.Package != nil { //nolint:nestif
		logE = logE.WithFields(logrus.Fields{
			"exe_path": which.ExePath,
			"package":  which.Package.Package.Name,
		})

		var checksums *checksum.Checksums
		if which.Config.ChecksumEnabled() {
			checksums = checksum.New()
			checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(ctrl.fs, which.ConfigFilePath)
			if err != nil {
				return err //nolint:wrapcheck
			}
			if err := checksums.ReadFile(ctrl.fs, checksumFilePath); err != nil {
				return fmt.Errorf("read a checksum JSON: %w", err)
			}
			defer func() {
				if err := checksums.UpdateFile(ctrl.fs, checksumFilePath); err != nil {
					logE.WithError(err).Error("update a checksum file")
				}
			}()
		}

		if err := ctrl.packageInstaller.InstallPackage(ctx, logE, which.Package, checksums); err != nil {
			return err //nolint:wrapcheck
		}
		for i := 0; i < 10; i++ {
			logE.Debug("check if exec file exists")
			if fi, err := ctrl.fs.Stat(which.ExePath); err == nil {
				if util.IsOwnerExecutable(fi.Mode()) {
					break
				}
			}
			logE.WithFields(logrus.Fields{
				"retry_count": i + 1,
			}).Debug("command isn't found. wait for lazy install")
			if err := util.Wait(ctx, 10*time.Millisecond); err != nil { //nolint:gomnd
				return fmt.Errorf("wait: %w", err)
			}
		}
	} else {
		logE = logE.WithFields(logrus.Fields{
			"exe_path": which.ExePath,
		})
	}

	p := filepath.Join(param.Dest, exeName)
	if ctrl.runtime.GOOS == "windows" && filepath.Ext(exeName) == "" {
		p += ".exe"
	}
	logE.WithFields(logrus.Fields{
		"exe_name": exeName,
		"dest":     p,
	}).Info("coping a file")
	dest, err := ctrl.fs.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePermission)
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	defer dest.Close()
	src, err := ctrl.fs.Open(which.ExePath)
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer src.Close()
	if _, err := io.Copy(dest, src); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}
