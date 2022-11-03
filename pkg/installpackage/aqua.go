package installpackage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) InstallAqua(ctx context.Context, logE *logrus.Entry, version string) error { //nolint:funlen
	assetTemplate := `aqua_{{.OS}}_{{.Arch}}.tar.gz`
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
			Checksum: &registry.Checksum{
				Type:       "github_release",
				Asset:      "aqua_{{trimV .Version}}_checksums.txt",
				FileFormat: "regexp",
				Algorithm:  "sha256",
				Pattern: &registry.ChecksumPattern{
					Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
					File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
				},
			},
		},
	}

	if err := inst.InstallPackage(ctx, logE, &domain.ParamInstallPackage{
		Checksums: checksum.New(),
		Pkg:       pkg,
	}); err != nil {
		return err
	}

	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
	})

	pkgPath, err := pkg.GetPkgPath(inst.rootDir, inst.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}
	fileSrc, err := pkg.GetFileSrc(&registry.File{
		Name: "aqua",
	}, inst.runtime)
	if err != nil {
		return fmt.Errorf("get a file path to aqua: %w", err)
	}

	p := filepath.Join(pkgPath, fileSrc)

	if inst.runtime.GOOS == "windows" {
		return inst.createAquaWindows(logE, p)
	}

	// create a symbolic link
	binName := "aqua"
	a, err := filepath.Rel(filepath.Join(inst.rootDir, "bin"), p)
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return inst.createLink(filepath.Join(inst.rootDir, "bin", binName), a, logE)
}

const (
	aquaBatTemplate = `@echo off
<AQUA> %*
`
	aquaScrTemplate = `#!/usr/bin/env bash
exec <AQUA> $@
`
)

func (inst *Installer) createAquaWindows(logE *logrus.Entry, p string) error {
	if err := inst.createBinWindows(filepath.Join(inst.rootDir, "bin", "aqua"), strings.Replace(aquaScrTemplate, "<AQUA>", p, 1), logE); err != nil {
		return err
	}
	if err := inst.createBinWindows(filepath.Join(inst.rootDir, "bat", "aqua.bat"), strings.Replace(aquaBatTemplate, "<AQUA>", p, 1), logE); err != nil {
		return err
	}
	return nil
}
