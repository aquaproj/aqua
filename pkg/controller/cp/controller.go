package cp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
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
	policyConfigReader policy.Reader
	policyConfigFinder policy.ConfigFinder
	requireChecksum    bool
}

func New(param *config.Param, pkgInstaller PackageInstaller, fs afero.Fs, rt *runtime.Runtime, whichCtrl which.Controller, installer Installer, policyConfigReader policy.Reader, policyConfigFinder policy.ConfigFinder) *Controller {
	return &Controller{
		rootDir:            param.RootDir,
		packageInstaller:   pkgInstaller,
		fs:                 fs,
		runtime:            rt,
		which:              whichCtrl,
		installer:          installer,
		policyConfigReader: policyConfigReader,
		policyConfigFinder: policyConfigFinder,
		requireChecksum:    param.RequireChecksum,
	}
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

var errCopyFailure = errors.New("it failed to copy some tools")

func (c *Controller) Copy(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if err := osfile.MkdirAll(c.fs, param.Dest); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}
	if len(param.Args) == 0 {
		return c.installer.Install(ctx, logE, param) //nolint:wrapcheck
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

	c.packageInstaller.SetCopyDir("")

	policyCfgs, err := c.policyConfigReader.Read(param.PolicyConfigFilePaths)
	if err != nil {
		return fmt.Errorf("read policy files: %w", err)
	}

	globalPolicyPaths := make(map[string]struct{}, len(param.PolicyConfigFilePaths))
	for _, p := range param.PolicyConfigFilePaths {
		globalPolicyPaths[p] = struct{}{}
	}

	for _, exeName := range param.Args {
		go func(exeName string) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			defer func() {
				<-maxInstallChan
			}()
			logE := logE.WithField("exe_name", exeName)
			if err := c.installAndCopy(ctx, logE, param, exeName, policyCfgs, globalPolicyPaths); err != nil {
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

func (c *Controller) installAndCopy(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string, policyConfigs []*policy.Config, globalPolicyPaths map[string]struct{}) error {
	findResult, err := c.which.Which(ctx, logE, param, exeName)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if findResult.Package != nil {
		logE = logE.WithField("package", findResult.Package.Package.Name)

		policyConfigs, err := c.policyConfigReader.Append(logE, findResult.ConfigFilePath, policyConfigs, globalPolicyPaths)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := c.install(ctx, logE, findResult, policyConfigs); err != nil {
			return err
		}
	}

	if err := c.copy(logE, param, findResult, exeName); err != nil {
		return err
	}
	return nil
}
