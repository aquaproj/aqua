package cp

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) install(ctx context.Context, logger *slog.Logger, findResult *which.FindResult, policyConfigs []*policy.Config, param *config.Param) error {
	checksums, updateChecksum, err := checksum.Open(
		logger, c.fs, findResult.ConfigFilePath,
		param.ChecksumEnabled(findResult.Config))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	if err := c.packageInstaller.InstallPackage(ctx, logger, &installpackage.ParamInstallPackage{
		Pkg:             findResult.Package,
		Checksums:       checksums,
		RequireChecksum: findResult.Config.RequireChecksum(param.EnforceRequireChecksum, param.RequireChecksum),
		ConfigFileDir:   filepath.Dir(findResult.ConfigFilePath),
		PolicyConfigs:   policyConfigs,
		DisablePolicy:   param.DisablePolicy,
	}); err != nil {
		return fmt.Errorf("install a package: %w", slogerr.With(err))
	}
	return c.packageInstaller.WaitExe(ctx, logger, findResult.ExePath) //nolint:wrapcheck
}
