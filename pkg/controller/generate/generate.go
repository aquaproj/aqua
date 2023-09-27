package generate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

// Generate searches packages in registries and outputs the configuration to standard output.
// If no package is specified, the interactive fuzzy finder is launched.
// If the package supports, the latest version is gotten by GitHub API.
func (c *Controller) Generate(ctx context.Context, logE *logrus.Entry, param *config.Param, args ...string) error {
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
	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return err //nolint:wrapcheck
	}

	cfg := &aqua.Config{}
	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	list, err := c.listPkgs(ctx, logE, param, cfg, cfgFilePath, args...)
	if err != nil {
		return err
	}
	list = excludeDuplicatedPkgs(logE, cfg, list)
	if len(list) == 0 {
		return nil
	}

	return c.outputter.Output(&output.Param{ //nolint:wrapcheck
		Insert:         param.Insert,
		Dest:           param.Dest,
		List:           list,
		ConfigFilePath: cfgFilePath,
	})
}

func (c *Controller) listPkgs(ctx context.Context, logE *logrus.Entry, param *config.Param, cfg *aqua.Config, cfgFilePath string, args ...string) ([]*aqua.Package, error) {
	var checksums *checksum.Checksums
	if cfg.ChecksumEnabled() {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(c.fs, cfgFilePath)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		if err := checksums.ReadFile(c.fs, checksumFilePath); err != nil {
			return nil, fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(c.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if param.File != "" || len(args) != 0 {
		return c.listPkgsWithoutFinder(ctx, logE, param, registryContents, args...)
	}

	return c.listPkgsWithFinder(ctx, logE, param, registryContents)
}

func (c *Controller) listPkgsWithFinder(ctx context.Context, logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config) ([]*aqua.Package, error) {
	// maps the package and the registry
	var pkgs []fuzzyfinder.Item
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			pkgs = append(pkgs, &fuzzyfinder.Package{
				PackageInfo:  pkg,
				RegistryName: registryName,
			})
		}
	}

	// Launch the fuzzy finder
	idxes, err := c.fuzzyFinder.FindMulti(pkgs, true)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil, nil
		}
		return nil, fmt.Errorf("find the package: %w", err)
	}
	arr := make([]*aqua.Package, len(idxes))
	for i, idx := range idxes {
		arr[i] = c.getOutputtedPkg(ctx, logE, param, pkgs[idx].(*fuzzyfinder.Package)) //nolint:forcetypeassert
	}

	return arr, nil
}

func (c *Controller) setPkgMap(logE *logrus.Entry, registryContents map[string]*registry.Config, m map[string]*fuzzyfinder.Package) {
	for registryName, registryContent := range registryContents {
		logE := logE.WithField("registry_name", registryName)
		for pkgName, pkg := range registryContent.PackageInfos.ToMap(logE) {
			pkg := pkg
			logE := logE.WithField("package_name", pkgName)
			m[registryName+","+pkgName] = &fuzzyfinder.Package{
				PackageInfo:  pkg,
				RegistryName: registryName,
			}
			for _, alias := range pkg.Aliases {
				if alias.Name == "" {
					logE.Warn("ignore a package alias because the alias is empty")
					continue
				}
				m[registryName+","+alias.Name] = &fuzzyfinder.Package{
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

func (c *Controller) listPkgsWithoutFinder(ctx context.Context, logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config, pkgNames ...string) ([]*aqua.Package, error) {
	m := map[string]*fuzzyfinder.Package{}
	c.setPkgMap(logE, registryContents, m)

	outputPkgs := []*aqua.Package{}
	for _, pkgName := range pkgNames {
		pkgName = getGeneratePkg(pkgName)
		key, version, _ := strings.Cut(pkgName, "@")
		findingPkg, ok := m[key]
		if !ok {
			return nil, logerr.WithFields(errUnknownPkg, logrus.Fields{"package_name": pkgName}) //nolint:wrapcheck
		}
		findingPkg.Version = version
		outputPkg := c.getOutputtedPkg(ctx, logE, param, findingPkg)
		outputPkgs = append(outputPkgs, outputPkg)
	}

	if param.File != "" {
		pkgs, err := c.readGeneratedPkgsFromFile(ctx, logE, param, outputPkgs, m)
		if err != nil {
			return nil, err
		}
		outputPkgs = pkgs
	}
	return outputPkgs, nil
}

func (c *Controller) getVersionFromGitHub(ctx context.Context, logE *logrus.Entry, param *config.Param, pkgInfo *registry.PackageInfo) string {
	if pkgInfo.VersionSource == "github_tag" {
		return c.getVersionFromGitHubTag(ctx, logE, param, pkgInfo)
	}
	if param.SelectVersion {
		return c.selectVersionFromReleases(ctx, logE, pkgInfo)
	}
	if pkgInfo.VersionFilter != "" || pkgInfo.VersionPrefix != "" {
		return c.listAndGetTagName(ctx, logE, pkgInfo)
	}
	return c.getVersionFromLatestRelease(ctx, logE, pkgInfo)
}

func (c *Controller) getVersion(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *fuzzyfinder.Package) string {
	if pkg.Version != "" {
		return pkg.Version
	}
	pkgInfo := pkg.PackageInfo
	if pkgInfo.Type == "cargo" {
		return c.getCargoVersion(ctx, logE, param, pkg)
	}
	if c.github == nil {
		return ""
	}
	if pkgInfo.HasRepo() {
		return c.getVersionFromGitHub(ctx, logE, param, pkgInfo)
	}
	return ""
}

func (c *Controller) getOutputtedPkg(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *fuzzyfinder.Package) *aqua.Package {
	outputPkg := &aqua.Package{
		Name:     pkg.PackageInfo.GetName(),
		Registry: pkg.RegistryName,
		Version:  pkg.Version,
	}
	if param.Detail {
		outputPkg.Link = pkg.PackageInfo.GetLink()
		outputPkg.Description = pkg.PackageInfo.Description
	}
	if outputPkg.Registry == registryStandard {
		outputPkg.Registry = ""
	}
	if outputPkg.Version == "" {
		version := c.getVersion(ctx, logE, param, pkg)
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
