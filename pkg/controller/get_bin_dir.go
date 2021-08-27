package controller

import (
	"fmt"
	"os"
	"path/filepath"
)

func (ctrl *Controller) GetBinDir(cfgFilePath string) error {
	cfg := &Config{}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	cfgFilePath = ctrl.getConfigFilePath(wd, cfgFilePath)
	if cfgFilePath == "" {
		return errConfigFileNotFound
	}
	if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
		return err
	}
	fmt.Fprintln(ctrl.Stdout, filepath.Join(filepath.Dir(cfgFilePath), ".aqua", "bin"))
	return nil
}
