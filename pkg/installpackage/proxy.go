package installpackage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

const ProxyVersion = "v1.2.14" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "aae4b9004237c6de27c4eef18cd07ea978d582ef18db400617ebfe5b9ad01659",
		"darwin/arm64":  "ef088450b7af7f44e064d9c4f4c9b40ac79166c379532234dda2cd9dff5961b6",
		"linux/amd64":   "d66234878fd9c5354d042108209cbf809b98f1f3601e5ff0c8d23d8198b68151",
		"linux/arm64":   "dcaeb1ed642798f8f72ac0f441bd72824535076b1b8666ca3fcc96a527fcdc80",
		"windows/amd64": "a7e716ad313f387d6305fba8a1d0404af21d9db3d82d5cb6ec1717d075b2792e",
		"windows/arm64": "00fc3454719dd2358e4ea0867e414f1623ff14fa2754bd598fa60119112f2981",
	}
}

func proxyPkg() *config.Package {
	return &config.Package{
		Package: &aqua.Package{
			Name:    proxyName,
			Version: ProxyVersion,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      pkgTypeGitHubRelease,
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
	finfo, err := os.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		chksum := ProxyChecksums()[is.runtime.Env()]
		if err := is.downloadWithRetry(ctx, logger, &DownloadParam{
			Package: pkg,
			Dest:    pkgPath,
			Asset:   assetName,
			Checksum: &checksum.Checksum{
				Algorithm: algoSHA256,
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
