package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/aquaproj/aqua/pkg/validate"
)

const proxyName = "aqua-proxy"

func (ctrl *Controller) Install(ctx context.Context, param *config.Param) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	rootBin := filepath.Join(ctrl.RootDir, "bin")

	if err := util.MkdirAll(rootBin); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}

	if err := ctrl.PackageInstaller.InstallProxy(ctx); err != nil {
		return err //nolint:wrapcheck
	}

	for _, cfgFilePath := range ctrl.ConfigFinder.Finds(wd, param.ConfigFilePath) {
		if err := ctrl.install(ctx, rootBin, cfgFilePath, param); err != nil {
			return err
		}
	}

	return ctrl.installAll(ctx, rootBin, param)
}

func (ctrl *Controller) installAll(ctx context.Context, rootBin string, param *config.Param) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range ctrl.ConfigFinder.GetGlobalConfigFilePaths() {
		if _, err := os.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := ctrl.install(ctx, rootBin, cfgFilePath, param); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) install(ctx context.Context, rootBin, cfgFilePath string, param *config.Param) error {
	cfg := &config.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := ctrl.ConfigReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	if err := validate.Config(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.RegistryInstaller.InstallRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return ctrl.PackageInstaller.InstallPackages(ctx, cfg, registryContents, rootBin, param.OnlyLink, param.IsTest) //nolint:wrapcheck
}
