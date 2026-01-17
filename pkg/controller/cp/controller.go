package cp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type Controller struct {
	packageInstaller PackageInstaller
	rootDir          string
	fs               afero.Fs
	runtime          *runtime.Runtime
	which            WhichController
	installer        Installer
	policyReader     PolicyReader
}

type PackageInstaller interface {
	InstallPackage(ctx context.Context, logger *slog.Logger, param *installpackage.ParamInstallPackage) error
	InstallPackages(ctx context.Context, logger *slog.Logger, param *installpackage.ParamInstallPackages) error
	SetCopyDir(copyDir string)
	Copy(dest, src string) error
	WaitExe(ctx context.Context, logger *slog.Logger, exePath string) error
}

type WhichController interface {
	Which(ctx context.Context, logger *slog.Logger, param *config.Param, exeName string) (*which.FindResult, error)
}

func New(param *config.Param, pkgInstaller PackageInstaller, fs afero.Fs, rt *runtime.Runtime, whichCtrl WhichController, installer Installer, policyReader PolicyReader) *Controller {
	return &Controller{
		rootDir:          param.RootDir,
		packageInstaller: pkgInstaller,
		fs:               fs,
		runtime:          rt,
		which:            whichCtrl,
		installer:        installer,
		policyReader:     policyReader,
	}
}

type PolicyReader interface {
	Read(policyFilePaths []string) ([]*policy.Config, error)
	Append(logger *slog.Logger, aquaYAMLPath string, policies []*policy.Config, globalPolicyPaths map[string]struct{}) ([]*policy.Config, error)
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

var errCopyFailure = errors.New("it failed to copy some tools")

func (c *Controller) Copy(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	if err := osfile.MkdirAll(c.fs, param.Dest); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}
	if len(param.Args) == 0 {
		return c.installer.Install(ctx, logger, param) //nolint:wrapcheck
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

	policyCfgs, err := c.policyReader.Read(param.PolicyConfigFilePaths)
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
			logger := logger.With("exe_name", exeName)
			if err := c.installAndCopy(ctx, logger, param, exeName, policyCfgs, globalPolicyPaths); err != nil {
				slogerr.WithError(logger, err).Error("install the package")
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

func (c *Controller) installAndCopy(ctx context.Context, logger *slog.Logger, param *config.Param, exeName string, policyConfigs []*policy.Config, globalPolicyPaths map[string]struct{}) error {
	findResult, err := c.which.Which(ctx, logger, param, exeName)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if findResult.Package != nil {
		logger := logger.With(
			"package_name", findResult.Package.Package.Name,
			"package_version", findResult.Package.Package.Version,
			"registry", findResult.Package.Package.Registry,
		)

		policyConfigs, err := c.policyReader.Append(logger, findResult.ConfigFilePath, policyConfigs, globalPolicyPaths)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := c.install(ctx, logger, findResult, policyConfigs, param); err != nil {
			return err
		}
	}

	if err := c.copy(logger, param, findResult.ExePath, exeName); err != nil {
		return err
	}
	return nil
}
