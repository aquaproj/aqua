package installpackage

import (
	"context"
	"fmt"
	"sync"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/cosign"
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
	f := false
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    "aquaproj/slsa-verifier",
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:               "github_release",
			RepoOwner:          "aquaproj",
			RepoName:           "slsa-verifier",
			Asset:              &assetTemplate,
			CompleteWindowsExt: &f,
		},
	}

	chksum := cosign.Checksums()[slsaVerifier.installer.runtime.Env()]

	pkgInfo, err := pkg.PackageInfo.Override(pkg.Package.Version, slsaVerifier.installer.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(slsaVerifier.installer.runtime, slsaVerifier.installer.runtime.Env())
	if err != nil {
		return fmt.Errorf("check if cosign is supported: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := slsaVerifier.installer.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums: checksum.New(), // Check cosign's checksum but not update aqua-checksums.json
		Pkg:       pkg,
		Checksum: &checksum.Checksum{
			Algorithm: "sha256",
			Checksum:  chksum,
		},
		// PolicyConfigs is nil, so the policy check is skipped
	}); err != nil {
		return err
	}

	return nil
}
