package updatechecksum

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) UpdateChecksum(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	for _, cfgFilePath := range c.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		if err := c.updateChecksum(ctx, logger, cfgFilePath); err != nil {
			return err
		}
	}

	return c.updateGlobalChecksumFiles(ctx, logger, param)
}

func (c *Controller) updateGlobalChecksumFiles(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := c.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := c.updateChecksum(ctx, logger, cfgFilePath); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) updateChecksum(ctx context.Context, logger *slog.Logger, cfgFilePath string) (namedErr error) { //nolint:cyclop
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	checksums := checksum.New()
	checksums.EnableOutput()
	checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(c.fs, cfgFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if err := checksums.ReadFile(c.fs, checksumFilePath); err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logger, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}
	pkgs, _ := config.ListPackagesNotOverride(logger, cfg, registryContents)
	failed := false
	defer func() {
		if c.prune {
			checksums.Prune()
		}
		if err := checksums.UpdateFile(c.fs, checksumFilePath); err != nil {
			namedErr = fmt.Errorf("update a checksum file: %w", err)
		}
	}()

	var supportedEnvs []string
	if cfg.Checksum != nil {
		supportedEnvs = cfg.Checksum.SupportedEnvs
	}

	for _, rgst := range cfg.Registries {
		if err := c.updateRegistry(ctx, logger, checksums, rgst); err != nil {
			failed = true
			slogerr.WithError(logger, err).Error("update checksums")
		}
	}

	for _, pkg := range pkgs {
		logger := logger.With(
			"package_name", pkg.Package.Name,
			"package_version", pkg.Package.Version,
			"package_registry", pkg.Package.Registry,
		)
		if err := c.updatePackage(ctx, logger, checksums, pkg, supportedEnvs); err != nil {
			failed = true
			slogerr.WithError(logger, err).Error("update checksums")
		}
	}
	if failed {
		return errFailedToUpdateChecksum
	}
	return nil
}

func (c *Controller) updateRegistry(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, rgst *aqua.Registry) error {
	if rgst.Type != "github_content" {
		return nil
	}
	rgstID := checksum.RegistryID(rgst)
	chksum := checksums.Get(rgstID)
	if chksum != nil {
		return nil
	}
	ghContentFile, err := c.registryDownloader.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
		RepoOwner: rgst.RepoOwner,
		RepoName:  rgst.RepoName,
		Ref:       rgst.Ref,
		Path:      rgst.Path,
	})
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer ghContentFile.Close()
	content := ghContentFile.Reader()
	algorithm := "sha256"
	chk, err := checksum.CalculateReader(content, algorithm)
	if err != nil {
		return fmt.Errorf("calculate a checksum of Registry: %w", err)
	}
	checksums.Set(rgstID, &checksum.Checksum{
		ID:        rgstID,
		Algorithm: algorithm,
		Checksum:  chk,
	})
	return nil
}

func (c *Controller) updatePackage(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, supportedEnvs []string) error {
	logger.Info("updating a package checksum")
	rts, err := checksum.GetRuntimesFromSupportedEnvs(supportedEnvs, pkg.PackageInfo.SupportedEnvs)
	if err != nil {
		return fmt.Errorf("get supported platforms: %w", err)
	}

	pkgs, assetNames, err := c.getPkgs(pkg, rts)
	if err != nil {
		return err
	}
	checksumFiles := map[string]struct{}{}
	for _, rt := range rts {
		env := rt.Env()
		logger := logger.With(
			"checksum_env", env,
		)
		pkg, ok := pkgs[env]
		if !ok {
			continue
		}
		if err := c.updatePackageByRuntime(ctx, logger, checksums, pkg, checksumFiles, rt, assetNames); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) getPkgs(pkg *config.Package, rts []*runtime.Runtime) (map[string]*config.Package, map[string]struct{}, error) {
	pkgs := make(map[string]*config.Package, len(rts))
	assets := make(map[string]struct{}, len(rts))
	for _, rt := range rts {
		env := rt.Env()
		pkgInfo := pkg.PackageInfo
		pkgInfo = pkgInfo.Copy()
		pkgInfo.OverrideByRuntime(rt)
		switch pkgInfo.Type {
		case "cargo", "go_install":
			// Skip updating checksums of these packages
			continue
		}
		pkgWithEnv := &config.Package{
			Package:     pkg.Package,
			PackageInfo: pkgInfo,
		}
		asset, err := pkgWithEnv.RenderAsset(rt)
		if err != nil {
			return nil, nil, fmt.Errorf("render an asset: %w", err)
		}
		assets[asset] = struct{}{}
		pkgs[env] = pkgWithEnv
	}
	return pkgs, assets, nil
}

func (c *Controller) getChecksums(ctx context.Context, logger *slog.Logger, pkg *config.Package, checksumFiles map[string]struct{}, rt *runtime.Runtime, assetNames map[string]struct{}, checksumID string) ([]*checksum.Checksum, error) { //nolint:funlen,cyclop
	if !pkg.PackageInfo.Checksum.GetEnabled() {
		cs, err := c.dlAssetAndGetChecksum(ctx, logger, pkg, rt)
		if err != nil {
			return nil, err
		}
		return []*checksum.Checksum{cs}, nil
	}

	checksumFileID, err := pkg.RenderChecksumFileID(rt)
	if err != nil {
		return nil, fmt.Errorf("render a checksum file ID: %w", err)
	}
	if _, ok := checksumFiles[checksumFileID]; ok {
		return nil, nil
	}
	checksumFiles[checksumFileID] = struct{}{}
	logger.Debug("downloading a checksum file")
	file, _, err := c.chkDL.DownloadChecksum(ctx, logger, rt, pkg)
	if err != nil {
		return nil, fmt.Errorf("download a checksum file: %w", err)
	}
	if file == nil {
		return nil, nil
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read a checksum file: %w", err)
	}
	return c.getChecksumsFromChecksumFile(pkg, assetNames, checksumID, strings.TrimSpace(string(b)))
}

func (c *Controller) getChecksumsFromChecksumFile(pkg *config.Package, assetNames map[string]struct{}, checksumID string, checksumFile string) ([]*checksum.Checksum, error) { //nolint:funlen,cyclop
	pkgInfo := pkg.PackageInfo
	m, s, err := checksum.ParseChecksumFile(checksumFile, pkgInfo.Checksum)
	if err != nil {
		return nil, fmt.Errorf("parse a checksum file: %w", err)
	}
	if s != "" {
		return []*checksum.Checksum{
			{
				ID:        checksumID,
				Checksum:  s,
				Algorithm: pkgInfo.Checksum.GetAlgorithm(),
			},
		}, nil
	}
	arr := make([]*checksum.Checksum, 0, len(m))
	for assetName, chksum := range m {
		if _, ok := assetNames[assetName]; !ok {
			continue
		}
		checksumID, err := pkg.ChecksumIDFromAsset(assetName)
		if err != nil {
			return nil, fmt.Errorf("get a checksum id from asset: %w", err)
		}
		arr = append(arr, &checksum.Checksum{
			ID:        checksumID,
			Checksum:  chksum,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		})
	}
	return arr, nil
}

func (c *Controller) updatePackageByRuntime(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, checksumFiles map[string]struct{}, rt *runtime.Runtime, assetNames map[string]struct{}) error { //nolint:funlen,cyclop
	checksumID, err := pkg.ChecksumID(rt)
	if err != nil {
		return fmt.Errorf("get a checksum id: %w", err)
	}

	if a := checksums.Get(checksumID); a != nil {
		return nil
	}

	cs, err := c.getChecksums(ctx, logger, pkg, checksumFiles, rt, assetNames, checksumID)
	if err != nil {
		return err
	}
	for _, c := range cs {
		if a := checksums.Get(c.ID); a != nil {
			continue
		}
		checksums.Set(c.ID, c)
	}
	return nil
}

func (c *Controller) dlAssetAndGetChecksum(ctx context.Context, logger *slog.Logger, pkg *config.Package, rt *runtime.Runtime) (*checksum.Checksum, error) {
	attrs := slogerr.NewAttrs(1)
	checksumID, err := pkg.ChecksumID(rt)
	if err != nil {
		return nil, fmt.Errorf("get a checksum id: %w", err)
	}
	assetName, err := pkg.RenderAsset(rt)
	if err != nil {
		return nil, fmt.Errorf("get an asset name: %w", err)
	}
	logger = attrs.Add(logger, "asset_name", assetName)
	logger.Info("downloading an asset to calculate the checksum")
	f, err := download.ConvertPackageToFile(pkg, assetName, rt)
	if err != nil {
		return nil, attrs.With(err) //nolint:wrapcheck
	}
	file, _, err := c.downloader.ReadCloser(ctx, logger, f)
	if err != nil {
		return nil, fmt.Errorf("download an asset: %w", attrs.With(err))
	}
	defer file.Close()
	algorithm := "sha256"
	chk, err := checksum.CalculateReader(file, algorithm)
	if err != nil {
		return nil, fmt.Errorf("calculate an asset: %w", slogerr.With(attrs.With(err), "algorithm", algorithm))
	}
	return &checksum.Checksum{
		ID:        checksumID,
		Checksum:  chk,
		Algorithm: algorithm,
	}, nil
}
