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

const ProxyVersion = "v1.2.9" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "427783de4e938644c95a37f9eef05338c6f52516bf2eca1aa121f50893515ea5",
		"darwin/arm64":  "a20af1cff23491a344c2d3f8ea25d196412e2eed7f7f9161cc70ad9d45c89327",
		"linux/amd64":   "645378f0f05383b2217117faf022dbaae53479d3f9f8522bb2ad256a2154516f",
		"linux/arm64":   "f6ee7d5441e5bbfc90ca9fce3ba4f7ded2c830d56f5edabba4f0340d31a029b4",
		"windows/amd64": "d5315a37c066958bdeb6c80d3ec791b4f6534cd3f2b8b0bb5941245c7ff95e70",
		"windows/arm64": "6ff53338fa368c6f1b6800ab18ebf0cc6a63a1ff735259ac1d9518146e8850e3",
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
		err := is.downloadWithRetry(ctx, logE, &DownloadParam{
			Package: pkg,
			Dest:    pkgPath,
			Asset:   assetName,
			Checksum: &checksum.Checksum{
				Algorithm: "sha256",
				Checksum:  chksum,
			},
		})
		if err != nil {
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
