package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) Install(ctx context.Context, param *Param) error { //nolint:cyclop
	cfg := &Config{}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	param.ConfigFilePath = ctrl.getConfigFilePath(wd, param.ConfigFilePath)
	if param.ConfigFilePath == "" {
		return errConfigFileNotFound
	}
	if err := ctrl.readConfig(param.ConfigFilePath, cfg); err != nil {
		return err
	}
	if cfg.BinDir == "" {
		cfg.BinDir = filepath.Join(filepath.Dir(param.ConfigFilePath), ".aqua", "bin")
	}
	if err := os.MkdirAll(cfg.BinDir, 0o775); err != nil { //nolint:gomnd
		return fmt.Errorf("create the directory: %w", err)
	}
	inlineRepo := make(map[string]*PackageInfo, len(cfg.InlineRepository))
	for _, pkgInfo := range cfg.InlineRepository {
		inlineRepo[pkgInfo.Name] = pkgInfo
	}

	if _, err := os.Stat(filepath.Join(ctrl.RootDir, "bin", "aqua-proxy")); err != nil {
		if err := ctrl.installProxy(ctx); err != nil {
			return err
		}
	}

	for _, pkg := range cfg.Packages {
		if err := ctrl.installPackage(ctx, inlineRepo, pkg, cfg); err != nil {
			return fmt.Errorf("install the package %s: %w", pkg.Name, err)
		}
	}
	return nil
}

func (ctrl *Controller) installPackage(ctx context.Context, inlineRepo map[string]*PackageInfo, pkg *Package, cfg *Config) error {
	logE := logrus.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"repository":      pkg.Repository,
	})
	logE.Info("install the package")
	if pkg.Repository != "inline" {
		return fmt.Errorf("only inline repository is supported (%s)", pkg.Repository)
	}
	pkgInfo, ok := inlineRepo[pkg.Name]
	if !ok {
		return fmt.Errorf("repository isn't found %s", pkg.Name)
	}

	assetName, err := pkgInfo.Artifact.Execute(map[string]interface{}{
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
	})
	if err != nil {
		return fmt.Errorf("render the asset name: %w", err)
	}

	pkgPath := getPkgPath(ctrl.RootDir, pkg, pkgInfo, assetName)
	// check if the repository exists
	logE.Debug("check if the package is already installed")
	finfo, err := os.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		if err := ctrl.download(ctx, pkg, pkgInfo, pkgPath, assetName); err != nil {
			return err
		}
	} else {
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	// create a symbolic link
	for _, file := range pkgInfo.Files {
		if _, err := os.Stat(filepath.Join(cfg.BinDir, file.Name)); err == nil {
			continue
		}
		if err := os.Symlink(filepath.Join(ctrl.RootDir, "bin", "aqua-proxy"), filepath.Join(cfg.BinDir, file.Name)); err != nil {
			return fmt.Errorf("create a symbolic link: %w", err)
		}
	}

	return nil
}

func getPkgPath(aquaRootDir string, pkg *Package, pkgInfo *PackageInfo, assetName string) string {
	return filepath.Join(aquaRootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.Repo, pkg.Version, assetName)
}

func (ctrl *Controller) installProxy(ctx context.Context) error {
	pkg := &Package{
		Name:       "aqua-proxy",
		Version:    "v0.1.0-0",
		Repository: "inline",
	}
	logE := logrus.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"repository":      pkg.Repository,
	})
	logE.Info("install the proxy")
	pkgInfo := &PackageInfo{
		Name:     "inline",
		Type:     "github_release",
		Repo:     "suzuki-shunsuke/aqua-proxy",
		Artifact: nil,
		Files: []*File{
			{
				Name: "aqua-proxy",
				Src:  "aqua-proxy",
			},
		},
	}

	assetName := "aqua-proxy_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"

	pkgPath := getPkgPath(ctrl.RootDir, pkg, pkgInfo, assetName)
	// check if the repository exists
	logE.Debug("check if the package is already installed")
	finfo, err := os.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		if err := ctrl.download(ctx, pkg, pkgInfo, pkgPath, assetName); err != nil {
			return err
		}
	} else {
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	// create a symbolic link
	if err := os.Symlink(filepath.Join(pkgPath, "aqua-proxy"), filepath.Join(ctrl.RootDir, "bin", "aqua-proxy")); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}

	return nil
}
