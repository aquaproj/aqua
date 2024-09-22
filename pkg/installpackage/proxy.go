package installpackage

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

const ProxyVersion = "v1.2.7" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "1EA8D0CD0A1BDCEE463B04DB7DF9DBF50AB54549B9EB9FE6506A15CDFA8F4563",
		"darwin/arm64":  "3686358F141D39A2909F9A6467E37B48DEF9767F7AF872483D43B4C9A5C5DF93",
		"linux/amd64":   "6917C867B818FA0B261E28C1924EFB33820D8FFF3930093D16B4E76793773F81",
		"linux/arm64":   "5E7B8A403E4B20251B02D0AEE537249C107DD6743DBD1F22A40EE6A077AC7DE9",
		"windows/amd64": "9C953DABD95A6231CF3C5C1A20781007AB44E603160C0865836DE57F3307976A",
		"windows/arm64": "4CBC6E4002EB5322A74E2B331B5BDAC188D4B678EF93DD2D2BED1AEF71683A8E",
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
		if isWindows(runtime.GOOS) {
			return is.recreateHardLinks()
		}
	} else { //nolint:gocritic
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	if isWindows(runtime.GOOS) {
		return nil
	}

	return is.createLink(logE, filepath.Join(is.rootDir, proxyName), a)
}
