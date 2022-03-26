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

	cfgFilePath, err := ctrl.ConfigFinder.Find(wd, param.ConfigFilePath)
	if err != nil {
		return err //nolint:wrapcheck
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
	for registryName, registryContent := range registryContents {
		for _, pkgInfo := range registryContent.PackageInfos {
			fmt.Fprintln(ctrl.Stdout, registryName+","+pkgInfo.GetName())
		}
	}

	return nil
}
