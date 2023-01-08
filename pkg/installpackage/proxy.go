package installpackage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

const ProxyVersion = "v1.1.4" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "9f7bfea8cfa38602194c6f92c1c4a0ad79fb54a1d7db08f446d3a78680bc8ea9",
		"darwin/arm64":  "b88992bf317af50109c32533b05e0bf19e1fb71489f18bbfe3ad3d1d0acee74b",
		"linux/amd64":   "902453d96fd1bd9a0053a86124663f49cf7ef50859b65aa8720dc868052a9762",
		"linux/arm64":   "3a4b8bb0665d0dc7e347711946c691d4ba85af51608310db2894d98b2c85694e",
		"windows/amd64": "7b9bec780b67b0f02face65f9bd63e69133b230a4af5bd71ff327dcb9bcbc57d",
		"windows/arm64": "f6e37900b7d7c0b189844be4a7597fa24781e89665862fa7cdf3b8ed38eec4c6",
	}
}

func (inst *InstallerImpl) InstallProxy(ctx context.Context, logE *logrus.Entry) error { //nolint:funlen
	if isWindows(inst.runtime.GOOS) {
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
	assetName, err := pkg.RenderAsset(inst.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}

	pkgPath, err := pkg.GetPkgPath(inst.rootDir, inst.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}
	logE.Debug("check if aqua-proxy is already installed")
	finfo, err := inst.fs.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		chksum := ProxyChecksums()[inst.runtime.Env()]
		if err := inst.downloadWithRetry(ctx, logE, &DownloadParam{
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
	a, err := filepath.Rel(filepath.Join(inst.rootDir, "bin"), filepath.Join(pkgPath, binName))
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return inst.createLink(filepath.Join(inst.rootDir, "bin", proxyName), a, logE)
}
