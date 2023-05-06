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

const ProxyVersion = "v1.2.0" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "8fb587a6b0667a1640a171cdf11a39d42c83d4bce2fdef43108916ee309b528c",
		"darwin/arm64":  "6b33d97d40357dc7b70732d475756a499b1216d1fcbe8a2bbb7fb07bba33b2c5",
		"linux/amd64":   "20827561aa4073d5fee84cff5733da72a60d113e7ed387d816b6d1e944c8041e",
		"linux/arm64":   "bbaee549e53b7b5e857458e01db2ecda60a787496238f2c8498ff8a5b933025a",
		"windows/amd64": "a5d8e05443b4708c271421f9c3e75a065913095c03c4cafe085be6cc81142479",
		"windows/arm64": "4c5cdf18a3fe702c3af7298690df90de6c76bfaa02e3a624acd1ebdacde7a887",
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
