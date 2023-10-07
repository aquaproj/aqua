package installpackage

import (
	"context"
	"fmt"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/sirupsen/logrus"
)

type SLSAVerifier struct {
	installer *InstallerImpl
	mutex     *sync.Mutex
}

func (sv *SLSAVerifier) installSLSAVerifier(ctx context.Context, logE *logrus.Entry, version string) error {
	sv.mutex.Lock()
	defer sv.mutex.Unlock()
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    "slsa-framework/slsa-verifier",
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "slsa-framework",
			RepoName:  "slsa-verifier",
			Asset:     "slsa-verifier-{{.OS}}-{{.Arch}}",
		},
	}

	chksum := slsa.Checksums()[sv.installer.runtime.Env()]

	pkgInfo, err := pkg.PackageInfo.Override(logE, pkg.Package.Version, sv.installer.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(sv.installer.runtime, sv.installer.runtime.Env())
	if err != nil {
		return fmt.Errorf("check if slsa-verifier is supported: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := sv.installer.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums: checksum.New(), // Check slsa-verifier's checksum but not update aqua-checksums.json
		Pkg:       pkg,
		Checksum: &checksum.Checksum{
			Algorithm: "sha512",
			Checksum:  chksum,
		},
		DisablePolicy: true,
	}); err != nil {
		return err
	}

	return nil
}
