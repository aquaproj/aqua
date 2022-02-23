package controller

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) Update(ctx context.Context, param *Param) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	for _, cfgFilePath := range ctrl.getConfigFilePaths(wd, param.ConfigFilePath) {
	}
	return nil
}

func (ctrl *Controller) update(ctx context.Context, rootBin, cfgFilePath string, param *Param) error {
	cfg := &Config{}
	if cfgFilePath == "" {
		return errConfigFileNotFound
	}
	if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
		return err
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.installRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return err
	}

	return ctrl.updatePackages(ctx, cfg, registryContents, rootBin, param.OnlyLink, param.IsTest)
}

func (ctrl *Controller) updatePackages(ctx context.Context, cfg *Config, registries map[string]*RegistryContent, binDir string) error { //nolint:funlen,cyclop,gocognit
	var failed bool
	for _, pkg := range cfg.Packages {
		logE := ctrl.logE().WithFields(logrus.Fields{
			"package_name":    pkg.Name,
			"package_version": pkg.Version,
			"registry":        pkg.Registry,
		})
		if registry, ok := cfg.Registries[pkg.Registry]; ok {
			if registry.Ref != "" {
				logE = logE.WithField("registry_ref", registry.Ref)
			}
		}
		pkgInfo, err := getPkgInfoFromRegistries(registries, pkg)
		if err != nil {
			logerr.WithError(logE, err).Error("update the package")
			failed = true
			continue
		}
		if !pkgInfo.HasRepo() {
			continue
		}
		tag := ctrl.getFilteredLatestTag(ctx, pkgInfo)
		if tag == "" || tag == pkg.Version {
			continue
		}
	}

	return nil
}
