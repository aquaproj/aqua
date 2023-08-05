package installpackage

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (inst *InstallerImpl) validatePackage(logE *logrus.Entry, param *ParamInstallPackage) error {
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo

	if err := pkgInfo.Validate(); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	if pkgInfo.IsNoAsset() {
		return errNoAsset
	}
	if pkgInfo.ErrorMessage != "" {
		logE.Error(pkgInfo.ErrorMessage)
		return errors.New("the package has a field `error_message`")
	}

	if !param.DisablePolicy {
		if err := inst.policyChecker.ValidatePackage(logE, param.Pkg, param.PolicyConfigs); err != nil {
			return err //nolint:wrapcheck
		}
	}

	if pkgInfo.Type == "go_install" && pkg.Package.Version == "latest" {
		return errGoInstallForbidLatest
	}
	return nil
}
