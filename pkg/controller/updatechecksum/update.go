package updatechecksum

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) UpdateChecksum(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	for _, cfgFilePath := range c.configFinder.Finds(param.CWD, param.ConfigFilePath) {
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
		if _, err := os.Stat(cfgFilePath); err != nil {
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
	checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(cfgFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if err := checksums.ReadFile(checksumFilePath); err != nil {
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
		if err := checksums.UpdateFile(checksumFilePath); err != nil {
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
		if err := c.checksumGetter.Get(ctx, logger, checksums, pkg, supportedEnvs); err != nil {
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
