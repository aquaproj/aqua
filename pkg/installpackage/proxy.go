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

const ProxyVersion = "v1.2.7-1" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "A2DC53DB2F8A272C586F0B84127E50DB0C27652BFF05ADD00B6777EC5889FB5F",
		"darwin/arm64":  "A5B21FD3CED16EC650E0772F3FB5EEFBFF2D45001C6DB314117F3290B461AB87",
		"linux/amd64":   "B1655278D54550B527B09FBED53F0684C10499C38FDB91AC1D7A91526D212861",
		"linux/arm64":   "68DD293008E8CCBDD7F0DC85B43CA87AB059BFA41EFDEEC6131A849CEFD6B99F",
		"windows/amd64": "FC9C5B4587A9AC33E2E38EC179C319744E95885358178462EC529134178479FA",
		"windows/arm64": "F2B2BC7E396407FC4DFABADE7A0D746D43681007D91733BCA69F15F67304CFE5",
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

	return is.createLink(filepath.Join(is.rootDir, proxyName), a, logE)
}
