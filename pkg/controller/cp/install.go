package cp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) install(ctx context.Context, logE *logrus.Entry, findResult *which.FindResult, policyConfigs []*policy.Config, param *config.Param) error {
	checksums, updateChecksum, err := checksum.Open(
		logE, c.fs, findResult.ConfigFilePath,
		param.ChecksumEnabled(findResult.Config))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	err := c.packageInstaller.InstallPackage(ctx, logE, &installpackage.ParamInstallPackage{
		Pkg:             findResult.Package,
		Checksums:       checksums,
		RequireChecksum: findResult.Config.RequireChecksum(param.EnforceRequireChecksum, param.RequireChecksum),
		ConfigFileDir:   filepath.Dir(findResult.ConfigFilePath),
		PolicyConfigs:   policyConfigs,
		DisablePolicy:   param.DisablePolicy,
	})
	if err != nil {
		return fmt.Errorf("install a package: %w", logerr.WithFields(err, logE.Data))
	}

	return c.packageInstaller.WaitExe(ctx, logE, findResult.ExePath) //nolint:wrapcheck
}
