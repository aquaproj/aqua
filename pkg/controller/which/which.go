package which

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type FindResult struct {
	Package        *config.Package
	File           *registry.File
	Config         *aqua.Config
	ExePath        string
	ConfigFilePath string
	EnableChecksum bool
}

func (c *Controller) Which(ctx context.Context, logger *slog.Logger, param *config.Param, exeName string) (*FindResult, error) {
	var filePaths []string
	if param.ConfigFilePath != "" {
		filePaths = []string{osfile.Abs(param.PWD, param.ConfigFilePath)}
	}
	for _, cfgFilePath := range append(filePaths, c.configFinder.Finds(param.PWD, "")...) {
		logger := logger.With("config_file_path", cfgFilePath)
		findResult, err := c.findExecFile(ctx, logger, param, cfgFilePath, exeName)
		if err != nil {
			return nil, err
		}
		if findResult != nil {
			return findResult, nil
		}
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		logger := logger.With("config_file_path", cfgFilePath)
		logger.Debug("checking a global configuration file")
		if _, err := c.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		findResult, err := c.findExecFile(ctx, logger, param, cfgFilePath, exeName)
		if err != nil {
			return nil, err
		}
		if findResult != nil {
			return findResult, nil
		}
	}

	if exePath := c.lookPath(c.osenv.Getenv("PATH"), exeName); exePath != "" {
		return &FindResult{
			ExePath: exePath,
		}, nil
	}
	return nil, slogerr.With(ErrCommandIsNotFound, //nolint:wrapcheck
		"exe_name", exeName,
		"doc", "https://aquaproj.github.io/docs/reference/codes/004",
	)
}

func (c *Controller) getExePath(findResult *FindResult) (string, error) {
	pkg := findResult.Package
	file := findResult.File
	if pkg.Package.Version == "" {
		return "", errVersionIsRequired
	}
	exePath, err := pkg.ExePath(c.rootDir, file, c.runtime)
	if err != nil {
		return exePath, err //nolint:wrapcheck
	}
	if file.Link == "" {
		return exePath, nil
	}
	return filepath.Join(filepath.Dir(exePath), file.Link), nil
}

func (c *Controller) findExecFile(ctx context.Context, logger *slog.Logger, param *config.Param, cfgFilePath, exeName string) (*FindResult, error) {
	cfg := &aqua.Config{}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return nil, err //nolint:wrapcheck
	}

	checksums, updateChecksum, err := checksum.Open(
		logger, c.fs, cfgFilePath,
		param.ChecksumEnabled(cfg))
	if err != nil {
		return nil, fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	logger.Debug("reading registry cache")
	registryCache, err := registry.NewCache(c.fs, param.RootDir, cfgFilePath)
	if err != nil {
		slogerr.WithError(logger, err).Debug("read a registry cache file", "config_file_path", cfgFilePath)
	}

	rgPaths := map[string]string{}
	registries := map[string]*registry.Config{}

	cacheKeys := map[string]map[string]struct{}{}
	if err := c.setRegistryCacheKeys(cfg, cfgFilePath, rgPaths, cacheKeys); err != nil {
		slogerr.WithError(logger, err).Warn("set registry cache keys")
	}

	defer func() {
		logger.Debug("updating registry cache")
		registryCache.Clean(cacheKeys)
		if err := registryCache.Write(); err != nil {
			slogerr.WithError(logger, err).Warn("write a registry cache file")
		}
	}()

	for _, pkg := range cfg.Packages {
		findResult, err := c.findExecFileFromPkg(ctx, logger, cfgFilePath, cfg, registryCache, rgPaths, registries, exeName, pkg, checksums)
		if err != nil {
			return nil, err
		}
		if findResult != nil {
			findResult.Config = cfg
			findResult.ConfigFilePath = cfgFilePath
			findResult.Package.Registry = cfg.Registries[pkg.Registry]
			return findResult, nil
		}
	}
	return nil, nil //nolint:nilnil
}

func (c *Controller) setRegistryCacheKeys(cfg *aqua.Config, cfgFilePath string, rgPaths map[string]string, cacheKeys map[string]map[string]struct{}) error {
	for _, pkg := range cfg.Packages {
		rg, ok := cfg.Registries[pkg.Registry]
		if !ok {
			continue
		}
		if rg.Type != aqua.RegistryTypeGitHubContent {
			continue
		}
		rgPath, ok := rgPaths[pkg.Registry]
		if !ok {
			p, err := rg.FilePath(c.rootDir, cfgFilePath)
			if err != nil {
				return fmt.Errorf("get a registry file path: %w", err)
			}
			rgPath = p
			rgPaths[pkg.Registry] = rgPath
		}
		keys, ok := cacheKeys[rgPath]
		if !ok {
			cacheKeys[rgPath] = map[string]struct{}{
				pkg.Name: {},
			}
			continue
		}
		keys[pkg.Name] = struct{}{}
	}
	return nil
}

func (c *Controller) findExecFileFromPkg(ctx context.Context, logger *slog.Logger, cfgFilePath string, cfg *aqua.Config, rCache *registry.Cache, rgPaths map[string]string, registries map[string]*registry.Config, exeName string, pkg *aqua.Package, checksums *checksum.Checksums) (*FindResult, error) { //nolint:cyclop
	if pkg.Registry == "" || pkg.Name == "" {
		logger.Debug("ignore a package because the package name or package registry name is empty")
		return nil, nil //nolint:nilnil
	}
	logger = logger.With(
		"registry_name", pkg.Registry,
		"package_name", pkg.Name,
	)
	pkgInfo, err := c.findPkgInfo(ctx, logger, cfgFilePath, cfg, rCache, rgPaths, registries, pkg, checksums)
	if err != nil {
		return nil, err
	}

	if pkgInfo == nil {
		logger.Warn("package isn't found")
		return nil, nil //nolint:nilnil
	}

	if !pkgInfo.MaybeHasCommand(exeName) && !pkg.HasCommandAlias(exeName) {
		return nil, nil //nolint:nilnil
	}

	pkgInfo, err = pkgInfo.Override(logger, pkg.Version, c.runtime)
	if err != nil {
		slogerr.WithError(logger, err).Warn("version constraint is invalid")
		return nil, nil //nolint:nilnil
	}

	supported, err := pkgInfo.CheckSupported(c.runtime, c.runtime.GOOS+"/"+c.runtime.GOARCH)
	if err != nil {
		slogerr.WithError(logger, err).Error("check if the package is supported")
		return nil, nil //nolint:nilnil
	}
	if !supported {
		logger.Debug("the package isn't supported on this environment")
		return nil, nil //nolint:nilnil
	}

	for _, file := range pkgInfo.GetFiles() {
		findResult, err := c.findExecFileFromFile(logger, exeName, pkg, pkgInfo, file)
		if err != nil {
			return nil, err
		}
		if findResult != nil {
			return findResult, nil
		}
	}
	return nil, nil //nolint:nilnil
}

func (c *Controller) findPkgInfo(ctx context.Context, logger *slog.Logger, cfgFilePath string, cfg *aqua.Config, rCache *registry.Cache, rgPaths map[string]string, registries map[string]*registry.Config, pkg *aqua.Package, checksums *checksum.Checksums) (*registry.PackageInfo, error) { //nolint:cyclop,funlen
	rg, ok := cfg.Registries[pkg.Registry]
	if !ok {
		logger.Debug("ignore a package because the registry isn't found")
		return nil, nil //nolint:nilnil
	}
	if err := rg.Validate(); err != nil {
		return nil, fmt.Errorf("validate the registry: %w", err)
	}
	if rg.Type != aqua.RegistryTypeGitHubContent {
		logger.Debug("getting a package from a registry")
		rc, ok := registries[pkg.Registry]
		if !ok {
			a, err := c.registryInstaller.InstallRegistry(ctx, logger, rg, cfgFilePath, checksums)
			if err != nil {
				return nil, fmt.Errorf("install a registry: %w", err)
			}
			rc = a
			registries[pkg.Registry] = rc
		}
		return rc.Package(logger, pkg.Name), nil
	}
	rgPath, ok := rgPaths[pkg.Registry]
	if !ok {
		p, err := rg.FilePath(c.rootDir, cfgFilePath)
		if err != nil {
			return nil, fmt.Errorf("get a registry file path: %w", err)
		}
		rgPath = p
		rgPaths[pkg.Registry] = rgPath
	}
	logger.Debug("getting a package from a registry cache",
		"registry_file_path", rgPath,
		"package_name", pkg.Name)
	if pkgInfo := rCache.Get(rgPath, pkg.Name); pkgInfo != nil {
		return pkgInfo, nil
	}
	logger.Debug("a package isn't found in a registry cache. Getting it from a registry")
	rc, ok := registries[pkg.Registry]
	if !ok {
		a, err := c.registryInstaller.InstallRegistry(ctx, logger, rg, cfgFilePath, checksums)
		if err != nil {
			return nil, fmt.Errorf("install a registry: %w", err)
		}
		rc = a
		registries[pkg.Registry] = rc
	}
	pkgInfo := rc.Package(logger, pkg.Name)
	if pkgInfo == nil {
		logger.Warn("package isn't found")
		return nil, nil //nolint:nilnil
	}
	logger.Debug("adding a package to the registry cache")
	rCache.Add(rgPath, pkgInfo)
	return pkgInfo, nil
}

func (c *Controller) findExecFileFromFile(logger *slog.Logger, exeName string, pkg *aqua.Package, pkgInfo *registry.PackageInfo, file *registry.File) (*FindResult, error) {
	cmds := map[string]struct{}{
		file.Name: {},
	}
	for _, alias := range pkg.CommandAliases {
		if file.Name != alias.Command {
			continue
		}
		cmds[alias.Alias] = struct{}{}
	}
	if _, ok := cmds[exeName]; !ok {
		return nil, nil //nolint:nilnil
	}
	findResult := &FindResult{
		Package: &config.Package{
			Package:     pkg,
			PackageInfo: pkgInfo,
		},
		File: file,
	}
	if err := findResult.Package.ApplyVars(); err != nil {
		return nil, fmt.Errorf("apply package variables: %w", err)
	}
	exePath, err := c.getExePath(findResult)
	if err != nil {
		slogerr.WithError(logger, err).Error("get the execution file path")
		return nil, nil //nolint:nilnil
	}
	findResult.ExePath = exePath
	return findResult, nil
}
