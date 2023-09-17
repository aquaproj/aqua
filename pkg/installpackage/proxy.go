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

const ProxyVersion = "v1.2.3" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "1ee85d8a99bd3b7605e18920b87fa6c5b5292fd84804df35ca86d45b595c1fbf",
		"darwin/arm64":  "e3d934fde6eb7c4af40c06426d375f6d2bb45cdbf9c233d14376403552345738",
		"linux/amd64":   "3c415f692befc7e432ac647c2e1c86921fce45997af371d7ba31f10404d427fc",
		"linux/arm64":   "263668268f68df81cccacec9f9f7faf7da0491750a35906331ecc5083bb3029a",
		"windows/amd64": "554bca6589b352d6dea48acd655bec8bf7025434e91e50cee677e605f85bfb75",
		"windows/arm64": "6df0e116448ee97f1f84a5cf2a6364bb92730e809a36b02485d5e4ba81b48ff1",
	}
}

func (is *InstallerImpl) InstallProxy(ctx context.Context, logE *logrus.Entry) error { //nolint:funlen
	if isWindows(is.runtime.GOOS) {
		return nil
	}
	proxyAssetTemplate := `aqua-proxy_{{.OS}}_{{.Arch}}.tar.gz`
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    proxyName,
			Version: ProxyVersion,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "aquaproj",
			RepoName:  proxyName,
			Asset:     &proxyAssetTemplate,
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
