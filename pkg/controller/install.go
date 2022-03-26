package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aquaproj/aqua/pkg/log"
	"github.com/sirupsen/logrus"
)

const proxyName = "aqua-proxy"

func (ctrl *Controller) Install(ctx context.Context, param *Param) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	rootBin := filepath.Join(ctrl.RootDir, "bin")

	if err := mkdirAll(rootBin); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}

	if err := ctrl.installProxy(ctx); err != nil {
		return err
	}

	for _, cfgFilePath := range ctrl.ConfigFinder.Finds(wd, param.ConfigFilePath) {
		if err := ctrl.install(ctx, rootBin, cfgFilePath, param); err != nil {
			return err
		}
	}

	return ctrl.installAll(ctx, rootBin, param)
}

func (ctrl *Controller) installAll(ctx context.Context, rootBin string, param *Param) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range getGlobalConfigFilePaths() {
		if _, err := os.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := ctrl.install(ctx, rootBin, cfgFilePath, param); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) install(ctx context.Context, rootBin, cfgFilePath string, param *Param) error {
	cfg := &Config{}
	if cfgFilePath == "" {
		return errConfigFileNotFound
	}
	if err := ctrl.ConfigReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.installRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return err
	}

	return ctrl.installPackages(ctx, cfg, registryContents, rootBin, param.OnlyLink, param.IsTest)
}

const defaultMaxParallelism = 5

func getMaxParallelism() int {
	envMaxParallelism := os.Getenv("AQUA_MAX_PARALLELISM")
	if envMaxParallelism == "" {
		return defaultMaxParallelism
	}
	num, err := strconv.Atoi(envMaxParallelism)
	if err != nil {
		log.New().WithFields(logrus.Fields{
			"AQUA_MAX_PARALLELISM": envMaxParallelism,
		}).Warn("the environment variable AQUA_MAX_PARALLELISM must be a number")
		return defaultMaxParallelism
	}
	if num <= 0 {
		return defaultMaxParallelism
	}
	return num
}
