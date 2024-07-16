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
		"darwin/amd64":  "b7bc0d22040bf450fbb31a8a92acabee11e33735e59a0692574ee57fbe4695be",
		"darwin/arm64":  "2f0a551c14cae1823d2cde2530348032ce04aff5bfe9b5d8fd5815a366d7a9e5",
		"linux/amd64":   "8fa0876fb9c6b23b9f911fa49ec8edc750badec86eb30b47e8aff231d82d12dd",
		"linux/arm64":   "4694cbbdf7e2a22778baf75c424db59c3810d795792234b72be1166d432786ef",
		"windows/amd64": "86ec4a91f4bee061482070ac1b6e519a37a27313372af259e367799f413d8dfd",
		"windows/arm64": "5e81237a2117b5f3f7a41c84181f722bfcb1032b7af8ff3977a94e1995dd0c47",
	}
}

func (is *Installer) InstallProxy(ctx context.Context, logE *logrus.Entry) error { //nolint:funlen
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

	// create a symbolic link
	binName := proxyName
	a, err := filepath.Rel(is.rootDir, filepath.Join(pkgPath, binName))
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	logE.Debug("check if aqua-proxy is already installed")
	finfo, err := is.fs.Stat(pkgPath)
	if err != nil { //nolint:nestif
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
		if isWindows(is.runtime.GOOS) {
			logE.Info("creating a hard link of aqua-proxy")
			if err := is.linker.Hardlink(a+".exe", filepath.Join(is.rootDir, proxyName+".exe")); err != nil {
				return fmt.Errorf("create a hard link of aqua-proxy: %w", err)
			}
			return nil
		}
	} else { //nolint:gocritic
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	return is.createLink(filepath.Join(is.rootDir, proxyName), a, logE)
}
