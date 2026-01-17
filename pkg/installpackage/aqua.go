package installpackage

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
)

func (is *Installer) InstallAqua(ctx context.Context, logger *slog.Logger, version string) error { //nolint:funlen
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

	pkgInfo, err := pkg.PackageInfo.Override(logger, pkg.Package.Version, is.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	pkg.PackageInfo = pkgInfo

	logger = logger.With(
		"package_name", pkg.Package.Name,
		"package_version", pkg.Package.Version,
		"registry", pkg.Package.Registry,
	)
	if err := is.InstallPackage(ctx, logger, &ParamInstallPackage{
		Checksums:     checksum.New(), // Check aqua's checksum but not update aqua-checksums.json
		Pkg:           pkg,
		DisablePolicy: true,
	}); err != nil {
		return err
	}

	exePath, err := pkg.ExePath(is.rootDir, &registry.File{
		Name: "aqua",
	}, is.runtime)
	if err != nil {
		return fmt.Errorf("get the executable file path: %w", err)
	}

	if is.runtime.GOOS == "windows" {
		return is.copyAquaOnWindows(exePath)
	}

	// create a symbolic link
	a, err := filepath.Rel(filepath.Join(is.rootDir, "bin"), exePath)
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return is.createLink(logger, filepath.Join(is.rootDir, "bin", "aqua"), a)
}

func (is *Installer) copyAquaOnWindows(exePath string) error {
	// https://github.com/orgs/aquaproj/discussions/2510
	// https://stackoverflow.com/questions/1211948/best-method-for-implementing-self-updating-software
	// https://github.com/aquaproj/aqua/issues/2918
	dest := filepath.Join(is.rootDir, "bin", "aqua.exe")
	if f, err := afero.Exists(is.fs, dest); err != nil {
		return fmt.Errorf("check if aqua.exe exists: %w", err)
	} else if f {
		// afero.Tempfile can't be used
		// > The system cannot move the file to a different disk drive
		tempDir := filepath.Join(is.rootDir, "temp")
		if err := osfile.MkdirAll(is.fs, tempDir); err != nil {
			return fmt.Errorf("create a temporary directory: %w", err)
		}
		if err := is.fs.Rename(dest, filepath.Join(tempDir, "aqua.exe")); err != nil {
			return fmt.Errorf("rename aqua.exe to update: %w", err)
		}
	}
	if err := is.linker.Hardlink(exePath, dest); err != nil {
		return fmt.Errorf("create a hard link to aqua.exe: %w", err)
	}
	return nil
}
