package installpackage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

func (inst *installer) InstallProxy(ctx context.Context, logE *logrus.Entry) error {
	proxyAssetTemplate := `aqua-proxy_{{.OS}}_{{.Arch}}.tar.gz`
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    proxyName,
			Version: "v1.1.0", // renovate: depName=aquaproj/aqua-proxy
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
		Type: inst.installers["github_release"],
	}
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})

	logE.Debug("check if aqua-proxy is already installed")
	f, err := pkg.Type.CheckInstalled(pkg)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if !f {
		// file doesn't exist
		logE.Debug("installing the proxy")
		if err := inst.downloadWithRetry(ctx, pkg, logE); err != nil {
			return err
		}
	}
	filePath, err := pkg.Type.GetFilePath(pkg, &registry.File{
		Name: proxyName,
	})
	if err != nil {
		return fmt.Errorf("get an install path of aqua-proxy: %w", err)
	}

	// create a symbolic link
	a, err := filepath.Rel(filepath.Join(inst.rootDir, "bin"), filePath)
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return inst.createLink(filepath.Join(inst.rootDir, "bin", proxyName), a, logE)
}
