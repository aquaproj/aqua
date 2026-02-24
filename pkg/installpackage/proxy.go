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

const ProxyVersion = "v1.2.13" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "7ea3fa8104c6a5dc502e9aef48bb041f80b06a7fba7f6d8a94b3ade830f25432",
		"darwin/arm64":  "c5b40aca37c859914cf74318d34ffa628039fd26cee45da6b2cd4c18daa4a79c",
		"linux/amd64":   "206c6ffdc0e5e243f836917594ffd96429f0c13ae859edd2bd6db32703f35142",
		"linux/arm64":   "937b391c3aef291a12822d56a7a7b86a8bae416a54043c6a1ea510a8b0cde935",
		"windows/amd64": "88de661cf7d65539f011549c51c8ddf773082ae037d9dd0a2c79ea1ea7d1e904",
		"windows/arm64": "faed3b38c2f07a1bb20381f5ab23b8de1e2ecd5c925ade65e4248d3ce3acb3e8",
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
