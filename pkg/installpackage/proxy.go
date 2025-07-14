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

const ProxyVersion = "v1.2.10" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "cbd424ab59f73bc8568f12eca3ddd9efc7e66f5ddb4b68e75ab5d0e69f364931",
		"darwin/arm64":  "28956e0778d299b42c4e415fb58311b495a79ca3e1efedacea171621b96aa442",
		"linux/amd64":   "488fb827c3e2d27237f19c7291ae4bb19a695d00de72105f9b307023fc8ac0d4",
		"linux/arm64":   "5f922f6c72ef52e217db9282190d301a117f944a66ef65b203c6d69b1c786da8",
		"windows/amd64": "6eef0b04d868890c1c406001c69144df676e43850dd8e72493ad7acd53f095cd",
		"windows/arm64": "dada7f8cb1fd2abe468a07988fa3d97f6570cca0704d24443f35e68b63e186eb",
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

func (is *Installer) InstallProxy(ctx context.Context, logE *logrus.Entry) error {
	pkg := proxyPkg()
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

	return is.createLink(logE, filepath.Join(is.rootDir, proxyName), a)
}
