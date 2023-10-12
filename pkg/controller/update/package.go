package update

import (
	"context"
	"errors"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

func (c *Controller) updatePackages(ctx context.Context, logE *logrus.Entry, cfgFilePath string, rgstCfgs map[string]*registry.Config) error {
	cfg := &aqua.Config{}
	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}
	pkgs, failed := config.ListPackages(logE, cfg, c.runtime, rgstCfgs)
	if len(pkgs) == 0 {
		if failed {
			return errors.New("list packages")
		}
		return nil
	}
	for _, pkg := range pkgs {
		logE := logE.WithFields(logrus.Fields{
			"package_name":    pkg.Package.Name,
			"package_version": pkg.Package.Version,
			"registry":        pkg.Package.Registry,
		})
		if err := c.updatePackage(ctx, logE, cfgFilePath, pkg); err != nil {
			return fmt.Errorf("update a package: %w", err)
		}
	}
	return nil
}

func (c *Controller) updatePackage(ctx context.Context, logE *logrus.Entry, cfgFilePath string, pkg *config.Package) error {
	// get a new version
	// update config file with AST
	return nil
}
