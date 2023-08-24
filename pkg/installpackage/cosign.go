package installpackage

import (
	"context"
	"fmt"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/sirupsen/logrus"
)

type Cosign struct {
	installer *InstallerImpl
	mutex     *sync.Mutex
}

func (c *Cosign) installCosign(ctx context.Context, logE *logrus.Entry, version string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	assetTemplate := `cosign-{{.OS}}-{{.Arch}}`
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    "sigstore/cosign",
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "sigstore",
			RepoName:  "cosign",
			Asset:     &assetTemplate,
			SupportedEnvs: []string{
				"darwin",
				"linux",
				"amd64",
			},
		},
	}

	chksum := cosign.Checksums()[c.installer.runtime.Env()]

	pkgInfo, err := pkg.PackageInfo.Override(logE, pkg.Package.Version, c.installer.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(c.installer.runtime, c.installer.runtime.Env())
	if err != nil {
		return fmt.Errorf("check if cosign is supported: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := c.installer.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums: checksum.New(), // Check cosign's checksum but not update aqua-checksums.json
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
