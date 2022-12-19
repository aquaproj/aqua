package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/policy"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const dirPermission os.FileMode = 0o775

type Controller struct {
	packageInstaller   domain.PackageInstaller
	rootDir            string
	configFinder       ConfigFinder
	configReader       domain.ConfigReader
	registryInstaller  domain.RegistryInstaller
	fs                 afero.Fs
	runtime            *runtime.Runtime
	skipLink           bool
	tags               map[string]struct{}
	excludedTags       map[string]struct{}
	policyConfigReader domain.PolicyConfigReader
	cosignInstaller    domain.CosignInstaller
}

func New(param *config.Param, configFinder ConfigFinder, configReader domain.ConfigReader, registInstaller domain.RegistryInstaller, pkgInstaller domain.PackageInstaller, fs afero.Fs, rt *runtime.Runtime, policyConfigReader domain.PolicyConfigReader, cosignInstaller domain.CosignInstaller) *Controller {
	return &Controller{
		rootDir:            param.RootDir,
		configFinder:       configFinder,
		configReader:       configReader,
		registryInstaller:  registInstaller,
		packageInstaller:   pkgInstaller,
		fs:                 fs,
		runtime:            rt,
		skipLink:           param.SkipLink,
		tags:               param.Tags,
		excludedTags:       param.ExcludedTags,
		policyConfigReader: policyConfigReader,
		cosignInstaller:    cosignInstaller,
	}
}

func (ctrl *Controller) Install(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if param.Dest == "" { //nolint:nestif
		rootBin := filepath.Join(ctrl.rootDir, "bin")
		if err := ctrl.fs.MkdirAll(rootBin, dirPermission); err != nil {
			return fmt.Errorf("create the directory: %w", err)
		}
		if ctrl.runtime.GOOS == "windows" {
			if err := ctrl.fs.MkdirAll(filepath.Join(ctrl.rootDir, "bat"), dirPermission); err != nil {
				return fmt.Errorf("create the directory: %w", err)
			}
		}

		if err := ctrl.packageInstaller.InstallProxy(ctx, logE); err != nil {
			return err //nolint:wrapcheck
		}
	}

	policyCfgs, err := ctrl.policyConfigReader.Read(param.PolicyConfigFilePaths)
	if err != nil {
		return fmt.Errorf("read policy files: %w", err)
	}

	if err := ctrl.cosignInstaller.InstallCosign(ctx, logE, cosign.Version); err != nil {
		return fmt.Errorf("install Cosign: %w", err)
	}

	for _, cfgFilePath := range ctrl.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		if err := ctrl.install(ctx, logE, cfgFilePath, policyCfgs); err != nil {
			return err
		}
	}

	return ctrl.installAll(ctx, logE, param, policyCfgs)
}

func (ctrl *Controller) installAll(ctx context.Context, logE *logrus.Entry, param *config.Param, policyConfigs []*policy.Config) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := ctrl.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := ctrl.install(ctx, logE, cfgFilePath, policyConfigs); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) install(ctx context.Context, logE *logrus.Entry, cfgFilePath string, policyConfigs []*policy.Config) error {
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return ctrl.packageInstaller.InstallPackages(ctx, logE, &domain.ParamInstallPackages{ //nolint:wrapcheck
		Config:         cfg,
		Registries:     registryContents,
		ConfigFilePath: cfgFilePath,
		SkipLink:       ctrl.skipLink,
		Tags:           ctrl.tags,
		ExcludedTags:   ctrl.excludedTags,
		PolicyConfigs:  policyConfigs,
	})
}
