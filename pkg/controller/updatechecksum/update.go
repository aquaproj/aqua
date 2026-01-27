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
		if err := c.getChecksum(ctx, logger, checksums, pkg, checksumFiles, rt, assetNames); err != nil {
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

func (c *Controller) getChecksum(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, checksumFiles map[string]struct{}, rt *runtime.Runtime, assetNames map[string]struct{}) error { //nolint:funlen,cyclop
	pkgInfo := pkg.PackageInfo

	if !pkg.PackageInfo.Checksum.GetEnabled() {
		if err := c.dlAssetAndGetChecksum(ctx, logger, checksums, pkg, rt); err != nil {
			return err
		}
		return nil
	}

	checksumID, err := pkg.ChecksumID(rt)
	if err != nil {
		return fmt.Errorf("get a checksum id: %w", err)
	}

	if a := checksums.Get(checksumID); a != nil {
		return nil
	}

	checksumFileID, err := pkg.RenderChecksumFileID(rt)
	if err != nil {
		return fmt.Errorf("render a checksum file ID: %w", err)
	}
	if _, ok := checksumFiles[checksumFileID]; ok {
		return nil
	}
	checksumFiles[checksumFileID] = struct{}{}
	logger.Debug("downloading a checksum file")
	file, _, err := c.chkDL.DownloadChecksum(ctx, logger, rt, pkg)
	if err != nil {
		return fmt.Errorf("download a checksum file: %w", err)
	}
	if file == nil {
		return nil
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read a checksum file: %w", err)
	}
	checksumFile := strings.TrimSpace(string(b))
	if pkgInfo.Checksum.FileFormat == "raw" {
		logger.Debug("set a checksum",
			"checksum_id", checksumID,
			"checksum", checksumFile)
		checksums.Set(checksumID, &checksum.Checksum{
			ID:        checksumID,
			Checksum:  checksumFile,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		})
		return nil
	}
	m, s, err := checksum.ParseChecksumFile(checksumFile, pkgInfo.Checksum)
	if err != nil {
		return fmt.Errorf("parse a checksum file: %w", err)
	}
	if s != "" {
		logger.Debug("set a checksum",
			"checksum_id", checksumID,
			"checksum", s)
		checksums.Set(checksumID, &checksum.Checksum{
			ID:        checksumID,
			Checksum:  s,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		})
		return nil
	}
	if len(m) == 1 {
		// get the asset name
		asset, err := pkg.RenderAsset(rt)
		if err != nil {
			return fmt.Errorf("render an asset: %w", err)
		}
		chksum, ok := m[asset]
		if !ok {
			// if the asset name is different, skip
			return nil
		}
		// update the checksum
		logger.Debug("set a checksum",
			"checksum_id", checksumID,
			"checksum", chksum)
		checksums.Set(checksumID, &checksum.Checksum{
			ID:        checksumID,
			Checksum:  chksum,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		})
		return nil
	}
	for assetName, chksum := range m {
		if _, ok := assetNames[assetName]; !ok {
			continue
		}
		checksumID, err := pkg.ChecksumIDFromAsset(assetName)
		if err != nil {
			return fmt.Errorf("get a checksum id from asset: %w", err)
		}
		logger.Debug("set a checksum",
			"checksum_id", checksumID,
			"checksum", chksum)
		checksums.Set(checksumID, &checksum.Checksum{
			ID:        checksumID,
			Checksum:  chksum,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		})
	}
	return nil
}

func (c *Controller) dlAssetAndGetChecksum(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, rt *runtime.Runtime) (gErr error) {
	attrs := slogerr.NewAttrs(1)
	defer func() {
		if gErr != nil {
			gErr = attrs.With(gErr)
		}
	}()
	checksumID, err := pkg.ChecksumID(rt)
	if err != nil {
		return fmt.Errorf("get a checksum id: %w", err)
	}
	if a := checksums.Get(checksumID); a != nil {
		return nil
	}
	assetName, err := pkg.RenderAsset(rt)
	if err != nil {
		return fmt.Errorf("get an asset name: %w", err)
	}
	logger = attrs.Add(logger, "asset_name", assetName)
	logger.Info("downloading an asset to calculate the checksum")
	f, err := download.ConvertPackageToFile(pkg, assetName, rt)
	if err != nil {
		return err //nolint:wrapcheck
	}
	file, _, err := c.downloader.ReadCloser(ctx, logger, f)
	if err != nil {
		return fmt.Errorf("download an asset: %w", err)
	}
	defer file.Close()
	algorithm := "sha256"
	chk, err := checksum.CalculateReader(file, algorithm)
	if err != nil {
		return fmt.Errorf("calculate an asset: %w", slogerr.With(err, "algorithm", algorithm))
	}
	checksums.Set(checksumID, &checksum.Checksum{
		ID:        checksumID,
		Checksum:  chk,
		Algorithm: algorithm,
	})
	return nil
}
