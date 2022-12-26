package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) InstallCosign(ctx context.Context, logE *logrus.Entry, version string) error {
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
			Checksum: &registry.Checksum{
				Type:       "github_release",
				Asset:      "cosign_checksums.txt",
				FileFormat: "regexp",
				Algorithm:  "sha256",
				Pattern: &registry.ChecksumPattern{
					Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
					File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
				},
			},
		},
	}

	pkgInfo, err := pkg.PackageInfo.Override(pkg.Package.Version, inst.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(inst.runtime, inst.runtime.GOOS+"/"+inst.runtime.GOARCH)
	if err != nil {
		return fmt.Errorf("check if cosign is supported: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := inst.InstallPackage(ctx, logE, &domain.ParamInstallPackage{
		Checksums: checksum.New(), // Check cosign's checksum but not update aqua-checksums.json
		Pkg:       pkg,
		// PolicyConfigs is nil, so the policy check is skipped
	}); err != nil {
		return err
	}

	return nil
}
