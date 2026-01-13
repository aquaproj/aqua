package generate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// Generate searches packages in registries and outputs the configuration to standard output.
// If no package is specified, the interactive fuzzy finder is launched.
// If the package supports, the latest version is gotten by GitHub API.
func (c *Controller) Generate(ctx context.Context, logger *slog.Logger, param *config.Param, args ...string) error { //nolint:cyclop
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
	cfgFilePath, err := c.getConfigFile(param)
	if err != nil {
		return err
	}

	cfg := &aqua.Config{}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	list, err := c.listPkgs(ctx, logger, param, cfg, cfgFilePath, args...)
	if err != nil {
		return err
	}
	list = excludeDuplicatedPkgs(logger, cfg, list)
	if len(list) == 0 {
		return nil
	}

	if cfg.ImportDir == "" {
		pkgs := make([]*aqua.Package, len(list))
		for i, pkg := range list {
			pkgs[i] = pkg.Package
		}
		return c.outputter.Output(&output.Param{ //nolint:wrapcheck
			Insert:         param.Insert,
			Dest:           param.Dest,
			List:           pkgs,
			ConfigFilePath: cfgFilePath,
		})
	}
	if param.Insert {
		if err := osfile.MkdirAll(c.fs, filepath.Join(filepath.Dir(cfgFilePath), cfg.ImportDir)); err != nil {
			return fmt.Errorf("create a directory specified by import_dir: %w", err)
		}
	}
	for _, pkg := range list {
		cmdName := pkg.PackageInfo.GetFiles()[0].Name
		dest := param.Dest
		if dest == "" && param.Insert {
			dest = filepath.Join(filepath.Dir(cfgFilePath), cfg.ImportDir, cmdName+".yaml")
		}
		if err := c.outputter.Output(&output.Param{
			Insert:         param.Insert,
			Dest:           dest,
			List:           []*aqua.Package{pkg.Package},
			ConfigFilePath: cfgFilePath,
		}); err != nil {
			return fmt.Errorf("output a package: %w", slogerr.With(err,
				"package_name", pkg.Package.Name,
			))
		}
	}
	return nil
}

func (c *Controller) getConfigFile(param *config.Param) (string, error) {
	if param.ConfigFilePath != "" || !param.Global {
		return c.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...) //nolint:wrapcheck
	}
	if len(param.GlobalConfigFilePaths) == 0 {
		return "", errors.New("no global configuration file is found")
	}
	return param.GlobalConfigFilePaths[0], nil
}

func (c *Controller) listPkgs(ctx context.Context, logger *slog.Logger, param *config.Param, cfg *aqua.Config, cfgFilePath string, args ...string) ([]*config.Package, error) {
	checksums, updateChecksum, err := checksum.Open(
		logger, c.fs, cfgFilePath, param.ChecksumEnabled(cfg))
	if err != nil {
		return nil, fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logger, cfg, cfgFilePath, checksums)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if param.File != "" || len(args) != 0 {
		return c.listPkgsWithoutFinder(ctx, logger, param, registryContents, args...)
	}

	return c.listPkgsWithFinder(ctx, logger, param, registryContents)
}

func (c *Controller) listPkgsWithFinder(ctx context.Context, logger *slog.Logger, param *config.Param, registryContents map[string]*registry.Config) ([]*config.Package, error) {
	// maps the package and the registry
	sizePkgs := 0
	for _, registryContent := range registryContents {
		sizePkgs += len(registryContent.PackageInfos)
	}
	items := make([]*fuzzyfinder.Item, 0, sizePkgs)
	pkgs := make([]*fuzzyfinder.Package, 0, sizePkgs)
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			p := &fuzzyfinder.Package{
				PackageInfo:  pkg,
				RegistryName: registryName,
			}
			pkgs = append(pkgs, p)
			items = append(items, &fuzzyfinder.Item{
				Item:    p.Item(),
				Preview: fuzzyfinder.PreviewPackage(p),
			})
		}
	}

	// Launch the fuzzy finder
	idxes, err := c.fuzzyFinder.FindMulti(items, true)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil, nil
		}
		return nil, fmt.Errorf("find the package: %w", err)
	}
	arr := make([]*config.Package, len(idxes))
	for i, idx := range idxes {
		arr[i] = c.getOutputtedPkg(ctx, logger, param, pkgs[idx])
	}

	return arr, nil
}

func (c *Controller) setPkgMap(logger *slog.Logger, registryContents map[string]*registry.Config, m map[string]*fuzzyfinder.Package) {
	for registryName, registryContent := range registryContents {
		logger := logger.With("registry_name", registryName)
		for pkgName, pkg := range registryContent.PackageInfos.ToMap(logger) {
			logger := logger.With("package_name", pkgName)
			m[registryName+","+pkgName] = &fuzzyfinder.Package{
				PackageInfo:  pkg,
				RegistryName: registryName,
			}
			for _, alias := range pkg.Aliases {
				if alias.Name == "" {
					logger.Warn("ignore a package alias because the alias is empty")
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

func (c *Controller) listPkgsWithoutFinder(ctx context.Context, logger *slog.Logger, param *config.Param, registryContents map[string]*registry.Config, pkgNames ...string) ([]*config.Package, error) {
	m := map[string]*fuzzyfinder.Package{}
	c.setPkgMap(logger, registryContents, m)

	outputPkgs := []*config.Package{}
	for _, pkgName := range pkgNames {
		pkgName = getGeneratePkg(pkgName)
		key, version, _ := strings.Cut(pkgName, "@")
		findingPkg, ok := m[key]
		if !ok {
			return nil, slogerr.With(errUnknownPkg, "package_name", pkgName) //nolint:wrapcheck
		}
		findingPkg.Version = version
		outputPkg := c.getOutputtedPkg(ctx, logger, param, findingPkg)
		outputPkgs = append(outputPkgs, outputPkg)
	}

	if param.File != "" {
		pkgs, err := c.readGeneratedPkgsFromFile(ctx, logger, param, outputPkgs, m)
		if err != nil {
			return nil, err
		}
		outputPkgs = pkgs
	}
	return outputPkgs, nil
}

func (c *Controller) getOutputtedPkg(ctx context.Context, logger *slog.Logger, param *config.Param, pkg *fuzzyfinder.Package) *config.Package {
	outputPkg := &config.Package{
		Package: &aqua.Package{
			Name:     pkg.PackageInfo.GetName(),
			Registry: pkg.RegistryName,
			Version:  pkg.Version,
		},
		PackageInfo: pkg.PackageInfo,
	}
	if param.Detail {
		outputPkg.Package.Link = pkg.PackageInfo.GetLink()
		outputPkg.Package.Description = pkg.PackageInfo.Description
	}
	if outputPkg.Package.Registry == registryStandard {
		outputPkg.Package.Registry = ""
	}
	if outputPkg.Package.Version == "" {
		version := c.fuzzyGetter.Get(ctx, logger, pkg.PackageInfo, "", param.SelectVersion, param.Limit)
		if version == "" {
			outputPkg.Package.Version = "[SET PACKAGE VERSION]"
			return outputPkg
		}
		outputPkg.Package.Version = version
	}
	if param.Pin {
		return outputPkg
	}
	outputPkg.Package.Name += "@" + outputPkg.Package.Version
	outputPkg.Package.Version = ""
	return outputPkg
}
