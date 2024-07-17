package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/sirupsen/logrus"
)

func (is *Installer) installMinisign(ctx context.Context, logE *logrus.Entry) error {
	pkg := minisign.Package()

	chksum := minisign.Checksums()[is.realRuntime.Env()]

	pkgInfo, err := pkg.PackageInfo.Override(logE, pkg.Package.Version, is.realRuntime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(is.realRuntime, is.realRuntime.Env())
	if err != nil {
		return fmt.Errorf("check if minisign is supported: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := is.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums: checksum.New(), // Check minisign's checksum but not update aqua-checksums.json
		Pkg:       pkg,
		Checksum: &checksum.Checksum{
			Algorithm: "sha256",
			Checksum:  chksum,
		},
		DisablePolicy: true,
	}); err != nil {
		return err
	}

	return nil
}
