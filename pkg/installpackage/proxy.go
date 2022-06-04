package installpackage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubrelease"
	"github.com/sirupsen/logrus"
)

func (inst *installer) InstallProxy(ctx context.Context, logE *logrus.Entry) error {
	pkg := &config.Package{
		Name:    proxyName,
		Version: "v1.1.0", // renovate: depName=aquaproj/aqua-proxy
	}
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})

	proxyAssetTemplate := `aqua-proxy_{{.OS}}_{{.Arch}}.tar.gz`
	logE.Debug("install the proxy")
	file := &config.File{
		Name: proxyName,
	}
	pkgInfo := &config.PackageInfo{
		Type:      "github_release",
		RepoOwner: "aquaproj",
		RepoName:  proxyName,
		Asset:     &proxyAssetTemplate,
		Files:     []*config.File{file},
	}

	pkgType := inst.installers[githubrelease.PkgType]

	logE.Debug("check if aqua-proxy is already installed")
	f, err := pkgType.CheckInstalled(pkg, pkgInfo)
	if err != nil {
		return fmt.Errorf("check if aqua-proxy is already installed: %w", err)
	}
	if !f {
		// file doesn't exist
		if err := inst.downloadWithRetry(ctx, pkg, pkgInfo, pkgType, logE); err != nil {
			return err
		}
	}

	filePath, err := pkgType.GetFilePath(pkg, pkgInfo, file)
	if err != nil {
		return fmt.Errorf("get a file path to aqua-proxy: %w", err)
	}

	// create a symbolic link
	a, err := filepath.Rel(filepath.Join(inst.rootDir, "bin"), filePath)
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return inst.createLink(filepath.Join(inst.rootDir, "bin", proxyName), a, logE)
}
