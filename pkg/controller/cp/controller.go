package cp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/policy"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	packageInstaller   PackageInstaller
	rootDir            string
	fs                 afero.Fs
	runtime            *runtime.Runtime
	which              which.Controller
	installer          Installer
	policyConfigReader policy.ConfigReader
	policyConfigFinder policy.ConfigFinder
	policyValidator    policy.Validator
	requireChecksum    bool
}

func New(param *config.Param, pkgInstaller PackageInstaller, fs afero.Fs, rt *runtime.Runtime, whichCtrl which.Controller, installer Installer, policyConfigReader policy.ConfigReader, policyConfigFinder policy.ConfigFinder, policyValidator policy.Validator) *Controller {
	return &Controller{
		rootDir:            param.RootDir,
		packageInstaller:   pkgInstaller,
		fs:                 fs,
		runtime:            rt,
		which:              whichCtrl,
		installer:          installer,
		policyConfigReader: policyConfigReader,
		policyConfigFinder: policyConfigFinder,
		policyValidator:    policyValidator,
		requireChecksum:    param.RequireChecksum,
	}
}

var errCopyFailure = errors.New("it failed to copy some tools")

func (ctrl *Controller) Copy(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if err := util.MkdirAll(ctrl.fs, param.Dest); err != nil {
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

	policyCfgs, err := ctrl.readPolicy(logE, param)
	if err != nil {
		return err
	}

	for _, exeName := range param.Args {
		go func(exeName string) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			defer func() {
				<-maxInstallChan
			}()
			logE := logE.WithField("exe_name", exeName)
			if err := ctrl.installAndCopy(ctx, logE, param, exeName, policyCfgs); err != nil {
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

func (ctrl *Controller) readPolicy(logE *logrus.Entry, param *config.Param) ([]*policy.Config, error) {
	policyFile, err := ctrl.policyConfigFinder.Find("", param.PWD)
	if err != nil {
		return nil, fmt.Errorf("find a policy file: %w", err)
	}
	if policyFile != "" {
		if err := ctrl.policyValidator.Validate(policyFile); err != nil {
			if err := ctrl.policyValidator.Warn(logE, policyFile); err != nil {
				logE.WithError(err).Warn("warn an disallowed policy file")
			}
		} else {
			param.PolicyConfigFilePaths = append(param.PolicyConfigFilePaths, policyFile)
		}
	}
	return ctrl.policyConfigReader.Read(param.PolicyConfigFilePaths, param.DisablePolicy) //nolint:wrapcheck
}

func (ctrl *Controller) installAndCopy(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string, policyConfigs []*policy.Config) error {
	findResult, err := ctrl.which.Which(ctx, logE, param, exeName)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if findResult.Package != nil {
		logE = logE.WithField("package", findResult.Package.Package.Name)
		if err := ctrl.install(ctx, logE, findResult, policyConfigs); err != nil {
			return err
		}
	}

	if err := ctrl.copy(logE, param, findResult, exeName); err != nil {
		return err
	}
	return nil
}
