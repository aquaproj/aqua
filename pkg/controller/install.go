package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/aqua/pkg/log"
)

const proxyName = "aqua-proxy"

func (ctrl *Controller) Install(ctx context.Context, param *Param) error {
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
	rootBin := filepath.Join(ctrl.RootDir, "bin")

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	if err := mkdirAll(rootBin); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(rootBin, proxyName)); err != nil {
		if err := ctrl.installProxy(ctx); err != nil {
			return err
		}
	}

	registryContents, err := ctrl.installRegistries(ctx, cfg, param.ConfigFilePath)
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
