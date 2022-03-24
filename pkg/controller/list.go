package controller

import (
	"context"
	"fmt"
	"os"
)

func (ctrl *Controller) List(ctx context.Context, param *Param, args []string) error {
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

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.installRegistries(ctx, cfg, param.ConfigFilePath)
	if err != nil {
		return err
	}
	for registryName, registryContent := range registryContents {
		for _, pkgInfo := range registryContent.PackageInfos {
			fmt.Fprintln(ctrl.Stdout, registryName+","+pkgInfo.GetName())
		}
	}

	return nil
}

func (ctrl *Controller) getFirstConfig(wd, cfgFilePath string) (string, error) {
	if cfgFilePath = ctrl.getConfigFilePath(wd, cfgFilePath); cfgFilePath != "" {
		return cfgFilePath, nil
	}
	for _, p := range getGlobalConfigFilePaths() {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		return p, nil
	}
	return "", errConfigFileNotFound
}
