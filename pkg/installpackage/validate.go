package installpackage

import (
	"errors"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
)

func (is *Installer) validatePackage(logE *logrus.Entry, param *ParamInstallPackage) error {
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo

	err := pkgInfo.Validate()
	if err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	if pkgInfo.NoAsset {
		return errNoAsset
	}

	if pkgInfo.ErrorMessage != "" {
		logE.Error(pkgInfo.ErrorMessage)
		return errors.New("the package has a field `error_message`")
	}

	if !param.DisablePolicy {
		err := policy.ValidatePackage(logE, pkg, param.PolicyConfigs)
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	if pkgInfo.Type == "go_install" && pkg.Package.Version == "latest" {
		return errGoInstallForbidLatest
	}

	return nil
}
