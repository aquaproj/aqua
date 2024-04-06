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

const ProxyVersion = "v1.2.6" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "d1739118a6cda5ac370049ed6970eead5e135b8a99fa5cce7dda5fa5b2faf949",
		"darwin/arm64":  "3160a1dc2c103acac51fdc76f3ceadc0c558f1b3e45b6fe2011f15f9766d3a69",
		"linux/amd64":   "a5129f58218229448c8407e45f776c17caa7e5f6bca36d1f66e15c00a8ba78a2",
		"linux/arm64":   "25302e2aa016d80ec3bf6aabcccb366fc7b3817b781c9b9d7bbceac05199171c",
		"windows/amd64": "e2a9e57997b5d4a4e9005a266906441d1474e9f55083d71e3598295615f826fc",
		"windows/arm64": "d48c65b9e961318cb7d38f2b496a64f7893a4ebc7503fd857d28352b2aec4468",
	}
}

func (is *Installer) InstallProxy(ctx context.Context, logE *logrus.Entry) error { //nolint:funlen
	if isWindows(is.runtime.GOOS) {
		return nil
	}
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    proxyName,
			Version: ProxyVersion,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "aquaproj",
			RepoName:  proxyName,
			Asset:     "aqua-proxy_{{.OS}}_{{.Arch}}.tar.gz",
			Files: []*registry.File{
				{
					Name: proxyName,
				},
			},
		},
	}
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})

	logE.Debug("install the proxy")
	assetName, err := pkg.RenderAsset(is.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}

	pkgPath, err := pkg.PkgPath(is.rootDir, is.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}
	logE.Debug("check if aqua-proxy is already installed")
	finfo, err := is.fs.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		chksum := ProxyChecksums()[is.runtime.Env()]
		if err := is.downloadWithRetry(ctx, logE, &DownloadParam{
			Package: pkg,
			Dest:    pkgPath,
			Asset:   assetName,
			Checksum: &checksum.Checksum{
				Algorithm: "sha256",
				Checksum:  chksum,
			},
		}); err != nil {
			return err
		}
	} else { //nolint:gocritic
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	// create a symbolic link
	binName := proxyName
	a, err := filepath.Rel(is.rootDir, filepath.Join(pkgPath, binName))
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return is.createLink(filepath.Join(is.rootDir, proxyName), a, logE)
}
