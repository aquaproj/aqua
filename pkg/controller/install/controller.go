package install

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

type Controller struct {
	packageInstaller  Installer
	rootDir           string
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller RegistryInstaller
	fs                afero.Fs
	runtime           *runtime.Runtime
	tags              map[string]struct{}
	excludedTags      map[string]struct{}
	policyReader      PolicyReader
	skipLink          bool
}

func New(param *config.Param, configFinder ConfigFinder, configReader ConfigReader, registryInstaller RegistryInstaller, pkgInstaller Installer, fs afero.Fs, rt *runtime.Runtime, policyReader PolicyReader) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registryInstaller,
		packageInstaller:  pkgInstaller,
		fs:                fs,
		runtime:           rt,
		skipLink:          param.SkipLink,
		tags:              param.Tags,
		excludedTags:      param.ExcludedTags,
		policyReader:      policyReader,
	}
}

type Installer interface {
	InstallPackage(ctx context.Context, logger *slog.Logger, param *installpackage.ParamInstallPackage) error
	InstallPackages(ctx context.Context, logger *slog.Logger, param *installpackage.ParamInstallPackages) error
	InstallProxy(ctx context.Context, logger *slog.Logger) error
}

type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
}

type PolicyReader interface {
	Read(policyFilePaths []string) ([]*policy.Config, error)
	Append(logger *slog.Logger, aquaYAMLPath string, policies []*policy.Config, globalPolicyPaths map[string]struct{}) ([]*policy.Config, error)
}

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}
