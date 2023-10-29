package install

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	packageInstaller   installpackage.Installer
	rootDir            string
	configFinder       ConfigFinder
	configReader       ConfigReader
	registryInstaller  RegistryInstaller
	fs                 afero.Fs
	runtime            *runtime.Runtime
	tags               map[string]struct{}
	excludedTags       map[string]struct{}
	policyConfigFinder policy.ConfigFinder
	policyConfigReader PolicyReader
	skipLink           bool
	requireChecksum    bool
}

func New(param *config.Param, configFinder ConfigFinder, configReader ConfigReader, registInstaller RegistryInstaller, pkgInstaller installpackage.Installer, fs afero.Fs, rt *runtime.Runtime, policyConfigReader PolicyReader, policyConfigFinder policy.ConfigFinder) *Controller {
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
		policyConfigFinder: policyConfigFinder,
		requireChecksum:    param.RequireChecksum,
	}
}

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}

type PolicyReader interface {
	Read(policyFilePaths []string) ([]*policy.Config, error)
	Append(logE *logrus.Entry, aquaYAMLPath string, policies []*policy.Config, globalPolicyPaths map[string]struct{}) ([]*policy.Config, error)
}

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}
