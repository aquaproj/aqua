package cp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/policy"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) install(ctx context.Context, logE *logrus.Entry, findResult *which.FindResult, policyConfigs []*policy.Config) error {
	var checksums *checksum.Checksums
	if findResult.Config.ChecksumEnabled() {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(ctrl.fs, findResult.ConfigFilePath)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := checksums.ReadFile(ctrl.fs, checksumFilePath); err != nil {
			return fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(ctrl.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}

	if err := ctrl.packageInstaller.InstallPackage(ctx, logE, &installpackage.ParamInstallPackage{
		Pkg:             findResult.Package,
		Checksums:       checksums,
		RequireChecksum: findResult.Config.RequireChecksum(),
		ConfigFileDir:   filepath.Dir(findResult.ConfigFilePath),
		PolicyConfigs:   policyConfigs,
	}); err != nil {
		return fmt.Errorf("install a package: %w", logerr.WithFields(err, logE.Data))
	}
	return ctrl.packageInstaller.WaitExe(ctx, logE, findResult.ExePath) //nolint:wrapcheck
}
