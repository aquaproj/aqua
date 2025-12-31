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
)

const ProxyVersion = "v1.2.12" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "0d99f58838e71794fed1b39bf19a3d65a2b8a6a7c6a19c3ed80d8978bcd44bff",
		"darwin/arm64":  "d152fd1258ce51c65322c169b109e050ab56f8ac0efe655156fdbcdac28848b7",
		"linux/amd64":   "738787d95b39c4af3af6dc8c9b7ea98ec639b0f76a37f7f4589ec5eeb8c8fec5",
		"linux/arm64":   "6a7c8df94f58a759c45490d3de5e668d1f9dab44f78a3030b0ef027c0bdc1c11",
		"windows/amd64": "a894ea24a36a774a95d6a4cba9d620575caf47236eb735c91de97240079e992e",
		"windows/arm64": "2e9674147d9686509bb86a0a318ab970b405dd5f8950a28e69ff576720cf0b32",
	}
}

func proxyPkg() *config.Package {
	return &config.Package{
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
}

func (is *Installer) InstallProxy(ctx context.Context, logger *slog.Logger) error {
	pkg := proxyPkg()
	logger = logger.With(
		"package_name", pkg.Package.Name,
		"package_version", pkg.Package.Version,
		"registry", pkg.Package.Registry,
	)

	logger.Debug("install the proxy")
	assetName, err := pkg.RenderAsset(is.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}

	pkgPath, err := pkg.AbsPkgPath(is.rootDir, is.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// create a symbolic link
	binName := proxyName
	a, err := filepath.Rel(is.rootDir, filepath.Join(pkgPath, binName))
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	logger.Debug("check if aqua-proxy is already installed")
	finfo, err := is.fs.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		chksum := ProxyChecksums()[is.runtime.Env()]
		if err := is.downloadWithRetry(ctx, logger, &DownloadParam{
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
		if is.realRuntime.IsWindows() {
			return is.recreateHardLinks()
		}
	} else { //nolint:gocritic
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	if is.realRuntime.IsWindows() {
		return nil
	}

	return is.createLink(logger, filepath.Join(is.rootDir, proxyName), a)
}
