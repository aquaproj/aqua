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

func (slsaVerifier *SLSAVerifier) installSLSAVerifier(ctx context.Context, logE *logrus.Entry, version string) error {
	slsaVerifier.mutex.Lock()
	defer slsaVerifier.mutex.Unlock()
	assetTemplate := `slsa-verifier-{{.OS}}-{{.Arch}}`
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    "slsa-framework/slsa-verifier",
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "slsa-framework",
			RepoName:  "slsa-verifier",
			Asset:     &assetTemplate,
		},
	}

	chksum := slsa.Checksums()[slsaVerifier.installer.runtime.Env()]

	pkgInfo, err := pkg.PackageInfo.Override(pkg.Package.Version, slsaVerifier.installer.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(slsaVerifier.installer.runtime, slsaVerifier.installer.runtime.Env())
	if err != nil {
		return fmt.Errorf("check if slsa-verifier is supported: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := slsaVerifier.installer.InstallPackage(ctx, logE, &ParamInstallPackage{
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
