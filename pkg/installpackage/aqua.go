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

func (is *InstallerImpl) InstallAqua(ctx context.Context, logE *logrus.Entry, version string) error { //nolint:funlen
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
			Asset:     "aqua_{{.OS}}_{{.Arch}}.{{.Format}}",
			Format:    "tar.gz",
			Overrides: []*registry.Override{
				{
					GOOS:   "windows",
					Format: "zip",
				},
			},
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
			VersionConstraints: `semver(">= 2.17.0")`,
			VersionOverrides: []*registry.VersionOverride{
				{
					VersionConstraints: `semver(">= 1.26.0")`,
					SLSAProvenance: &registry.SLSAProvenance{
						Enabled: &disabled,
					},
					Overrides: []*registry.Override{},
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
				{
					VersionConstraints: `semver("< 1.26.0")`,
					Overrides:          []*registry.Override{},
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

	pkgInfo, err := pkg.PackageInfo.Override(logE, pkg.Package.Version, is.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	pkg.PackageInfo = pkgInfo

	if err := is.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums:     checksum.New(), // Check aqua's checksum but not update aqua-checksums.json
		Pkg:           pkg,
		DisablePolicy: true,
	}); err != nil {
		return err
	}

	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
	})

	exePath, err := pkg.ExePath(is.rootDir, &registry.File{
		Name: "aqua",
	}, is.runtime)
	if err != nil {
		return fmt.Errorf("get the executable file path: %w", err)
	}

	if is.runtime.GOOS == "windows" {
		return is.Copy(filepath.Join(is.rootDir, "bin", "aqua.exe"), exePath)
	}

	// create a symbolic link
	a, err := filepath.Rel(filepath.Join(is.rootDir, "bin"), exePath)
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return is.createLink(filepath.Join(is.rootDir, "bin", "aqua"), a, logE)
}
