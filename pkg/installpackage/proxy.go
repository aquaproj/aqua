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

const ProxyVersion = "v1.1.4" // renovate: depName=aquaproj/aqua-proxy

func ProxyChecksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "198be9c4e00175731e01724172ad3e2c141c4be17de702d5352babb11a446eb0",
		"darwin/arm64":  "0e0c421e6e17cde8c3c1726e22adb99a95caa799ecb7478cc4d3e56a9ae55884",
		"linux/amd64":   "1d5a9cc5cdab91d3527cca1a6dbafc9530deb9559a542b8a6137627a729ade20",
		"linux/arm64":   "02ee1dbf7d20b75422bde16a82f602e5f7318177c22a67ef50d83bffdc198018",
		"windows/amd64": "27ac08a887f3f873e500f96b3ee2459848fa0be5837af3f9051044afc9d173b3",
		"windows/arm64": "4ad911f0a982a4287b93922642818fa3afeead236c9f1403638bb01507a32eb6",
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

	pkgPath, err := pkg.GetPkgPath(filepath.Join(inst.rootDir, "internal"), inst.runtime)
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
