package generate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	rgst "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	stdin             io.Reader
	github            RepositoriesService
	registryInstaller rgst.Installer
	configFinder      ConfigFinder
	configReader      reader.ConfigReader
	fuzzyFinder       FuzzyFinder
	versionSelector   VersionSelector
	fs                afero.Fs
	outputter         Outputter
	cosignInstaller   domain.CosignInstaller
}

func New(configFinder ConfigFinder, configReader reader.ConfigReader, registInstaller rgst.Installer, gh RepositoriesService, fs afero.Fs, fuzzyFinder FuzzyFinder, versionSelector VersionSelector, cosignInstaller domain.CosignInstaller) *Controller {
	return &Controller{
		stdin:             os.Stdin,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		github:            gh,
		fs:                fs,
		fuzzyFinder:       fuzzyFinder,
		versionSelector:   versionSelector,
		outputter:         output.New(os.Stdout, fs),
		cosignInstaller:   cosignInstaller,
	}
}

// Generate searches packages in registries and outputs the configuration to standard output.
// If no package is specified, the interactive fuzzy finder is launched.
// If the package supports, the latest version is gotten by GitHub API.
func (ctrl *Controller) Generate(ctx context.Context, logE *logrus.Entry, param *config.Param, args ...string) error {
	// Find and read a configuration file (aqua.yaml).
	// Install registries
	// List outputted packages
	//   Get packages by fuzzy finder or from file or from arguments
	//   Get versions
	//     Get versions from arguments or GitHub API (GitHub Tag or GitHub Release) or fuzzy finder (-s)
	// Output packages
	//   Format outputs
	//     registry:
	//       omit standard registry
	//     version:
	//       merge version with package name
	//       set default value
	//   Output to Stdout or Update aqua.yaml (-i)
	cfgFilePath, err := ctrl.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return err //nolint:wrapcheck
	}

	cfg := &aqua.Config{}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	if err := ctrl.cosignInstaller.InstallCosign(ctx, logE, cosign.Version); err != nil {
		return fmt.Errorf("install Cosign: %w", err)
	}

	list, err := ctrl.listPkgs(ctx, logE, param, cfg, cfgFilePath, args...)
	if err != nil {
		return err
	}
	list = excludeDuplicatedPkgs(logE, cfg, list)
	if len(list) == 0 {
		return nil
	}

	return ctrl.outputter.Output(&output.Param{ //nolint:wrapcheck
		Insert:         param.Insert,
		Dest:           param.Dest,
		List:           list,
		ConfigFilePath: cfgFilePath,
	})
}

type FindingPackage struct {
	PackageInfo  *registry.PackageInfo
	RegistryName string
	Version      string
}

func (ctrl *Controller) listPkgs(ctx context.Context, logE *logrus.Entry, param *config.Param, cfg *aqua.Config, cfgFilePath string, args ...string) ([]*aqua.Package, error) {
	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if param.File != "" || len(args) != 0 {
		return ctrl.listPkgsWithoutFinder(ctx, logE, param, registryContents, args...)
	}

	return ctrl.listPkgsWithFinder(ctx, logE, param, registryContents)
}

func (ctrl *Controller) listPkgsWithFinder(ctx context.Context, logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config) ([]*aqua.Package, error) {
	// maps the package and the registry
	var pkgs []*FindingPackage
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			pkgs = append(pkgs, &FindingPackage{
				PackageInfo:  pkg,
				RegistryName: registryName,
			})
		}
	}

	// Launch the fuzzy finder
	idxes, err := ctrl.fuzzyFinder.Find(pkgs)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil, nil
		}
		return nil, fmt.Errorf("find the package: %w", err)
	}
	arr := make([]*aqua.Package, len(idxes))
	for i, idx := range idxes {
		arr[i] = ctrl.getOutputtedPkg(ctx, logE, param, pkgs[idx])
	}

	return arr, nil
}

func (ctrl *Controller) setPkgMap(logE *logrus.Entry, registryContents map[string]*registry.Config, m map[string]*FindingPackage) {
	for registryName, registryContent := range registryContents {
		logE := logE.WithField("registry_name", registryName)
		for pkgName, pkg := range registryContent.PackageInfos.ToMapWarn(logE) {
			pkg := pkg
			logE := logE.WithField("package_name", pkgName)
			m[registryName+","+pkgName] = &FindingPackage{
				PackageInfo:  pkg,
				RegistryName: registryName,
			}
			for _, alias := range pkg.Aliases {
				if alias.Name == "" {
					logE.Warn("ignore a package alias because the alias is empty")
					continue
				}
				m[registryName+","+alias.Name] = &FindingPackage{
					PackageInfo:  pkg,
					RegistryName: registryName,
				}
			}
		}
	}
}

func getGeneratePkg(s string) string {
	if !strings.Contains(s, ",") {
		return "standard," + s
	}
	return s
}

func (ctrl *Controller) listPkgsWithoutFinder(ctx context.Context, logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config, pkgNames ...string) ([]*aqua.Package, error) {
	m := map[string]*FindingPackage{}
	ctrl.setPkgMap(logE, registryContents, m)

	outputPkgs := []*aqua.Package{}
	for _, pkgName := range pkgNames {
		pkgName = getGeneratePkg(pkgName)
		key, version, _ := strings.Cut(pkgName, "@")
		findingPkg, ok := m[key]
		if !ok {
			return nil, logerr.WithFields(errUnknownPkg, logrus.Fields{"package_name": pkgName}) //nolint:wrapcheck
		}
		findingPkg.Version = version
		outputPkg := ctrl.getOutputtedPkg(ctx, logE, param, findingPkg)
		outputPkgs = append(outputPkgs, outputPkg)
	}

	if param.File != "" {
		pkgs, err := ctrl.readGeneratedPkgsFromFile(ctx, logE, param, outputPkgs, m)
		if err != nil {
			return nil, err
		}
		outputPkgs = pkgs
	}
	return outputPkgs, nil
}

func (ctrl *Controller) getVersionFromGitHub(ctx context.Context, logE *logrus.Entry, param *config.Param, pkgInfo *registry.PackageInfo) string {
	if pkgInfo.VersionSource == "github_tag" {
		return ctrl.getVersionFromGitHubTag(ctx, logE, param, pkgInfo)
	}
	if param.SelectVersion {
		return ctrl.selectVersionFromReleases(ctx, logE, pkgInfo)
	}
	if pkgInfo.VersionFilter != nil {
		return ctrl.listAndGetTagName(ctx, logE, pkgInfo)
	}
	return ctrl.getVersionFromLatestRelease(ctx, logE, pkgInfo)
}

func (ctrl *Controller) getVersion(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *FindingPackage) string {
	if pkg.Version != "" {
		return pkg.Version
	}
	if ctrl.github == nil {
		return ""
	}
	pkgInfo := pkg.PackageInfo
	if pkgInfo.HasRepo() {
		return ctrl.getVersionFromGitHub(ctx, logE, param, pkgInfo)
	}
	return ""
}

func (ctrl *Controller) getOutputtedPkg(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *FindingPackage) *aqua.Package {
	outputPkg := &aqua.Package{
		Name:     pkg.PackageInfo.GetName(),
		Registry: pkg.RegistryName,
		Version:  pkg.Version,
	}
	if outputPkg.Registry == registryStandard {
		outputPkg.Registry = ""
	}
	if outputPkg.Version == "" {
		version := ctrl.getVersion(ctx, logE, param, pkg)
		if version == "" {
			outputPkg.Version = "[SET PACKAGE VERSION]"
			return outputPkg
		}
		outputPkg.Version = version
	}
	if param.Pin {
		return outputPkg
	}
	outputPkg.Name += "@" + outputPkg.Version
	outputPkg.Version = ""
	return outputPkg
}
