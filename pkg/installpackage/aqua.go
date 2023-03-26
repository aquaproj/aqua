package installpackage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

func (inst *InstallerImpl) InstallAqua(ctx context.Context, logE *logrus.Entry, version string) error { //nolint:funlen
	assetTemplate := `aqua_{{.OS}}_{{.Arch}}.tar.gz`
	provTemplate := "multiple.intoto.jsonl"
	disabled := false
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    "aquaproj/aqua",
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "aquaproj",
			RepoName:  "aqua",
			Asset:     &assetTemplate,
			Files: []*registry.File{
				{
					Name: "aqua",
				},
			},
			SLSAProvenance: &registry.SLSAProvenance{
				Type:  "github_release",
				Asset: &provTemplate,
			},
			// Checksum: &registry.Checksum{
			// 	Type:       "github_release",
			// 	Asset:      "aqua_{{trimV .Version}}_checksums.txt",
			// 	FileFormat: "regexp",
			// 	Algorithm:  "sha256",
			// 	Pattern: &registry.ChecksumPattern{
			// 		Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
			// 		File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
			// 	},
			// 	Cosign: &registry.Cosign{
			// 		CosignExperimental: true,
			// 		Opts: []string{
			// 			"--signature",
			// 			"https://github.com/aquaproj/aqua/releases/download/{{.Version}}/aqua_{{trimV .Version}}_checksums.txt.sig",
			// 			"--certificate",
			// 			"https://github.com/aquaproj/aqua/releases/download/{{.Version}}/aqua_{{trimV .Version}}_checksums.txt.pem",
			// 		},
			// 	},
			// },
			VersionConstraints: `semver(">= 1.26.0")`,
			VersionOverrides: []*registry.VersionOverride{
				{
					VersionConstraints: "true",
					SLSAProvenance: &registry.SLSAProvenance{
						Enabled: &disabled,
					},
					Checksum: &registry.Checksum{
						Type:       "github_release",
						Asset:      "aqua_{{trimV .Version}}_checksums.txt",
						FileFormat: "regexp",
						Algorithm:  "sha256",
						Pattern: &registry.ChecksumPattern{
							Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
							File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
						},
						Cosign: &registry.Cosign{
							Opts: []string{},
						},
					},
				},
			},
		},
	}

	pkgInfo, err := pkg.PackageInfo.Override(pkg.Package.Version, inst.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	pkg.PackageInfo = pkgInfo

	if err := inst.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums: checksum.New(), // Check aqua's checksum but not update aqua-checksums.json
		Pkg:       pkg,
		// PolicyConfigs is nil, so the policy check is skipped
	}); err != nil {
		return err
	}

	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
	})

	exePath, err := inst.getExePath(pkg)
	if err != nil {
		return err
	}

	if inst.runtime.GOOS == "windows" {
		return inst.Copy(filepath.Join(inst.rootDir, "bin", "aqua.exe"), exePath)
	}

	// create a symbolic link
	a, err := filepath.Rel(filepath.Join(inst.rootDir, "bin"), exePath)
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return inst.createLink(filepath.Join(inst.rootDir, "bin", "aqua"), a, logE)
}

func (inst *InstallerImpl) getExePath(pkg *config.Package) (string, error) {
	pkgPath, err := pkg.GetPkgPath(inst.rootDir, inst.runtime)
	if err != nil {
		return "", err //nolint:wrapcheck
	}
	fileSrc, err := pkg.GetFileSrc(&registry.File{
		Name: "aqua",
	}, inst.runtime)
	if err != nil {
		return "", fmt.Errorf("get a file path to aqua: %w", err)
	}
	return filepath.Join(pkgPath, fileSrc), nil
}
