package update

import (
	"context"
	"errors"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/update/ast"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (c *Controller) updatePackages(ctx context.Context, logE *logrus.Entry, cfgFilePath string, rgstCfgs map[string]*registry.Config) error {
	cfg := &aqua.Config{}
	cfgs, err := c.configReader.ReadToUpdate(cfgFilePath, cfg)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}
	cfgs[cfgFilePath] = cfg
	newVersions := map[string]string{}
	for cfgPath, cfg := range cfgs {
		pkgs, failed := config.ListPackages(logE, cfg, c.runtime, rgstCfgs)
		if len(pkgs) == 0 {
			if failed {
				return errors.New("list packages")
			}
			continue
		}
		for _, pkg := range pkgs {
			logE := logE.WithFields(logrus.Fields{
				"package_name":    pkg.Package.Name,
				"package_version": pkg.Package.Version,
				"registry":        pkg.Package.Registry,
			})
			newVersion := c.fuzzyGetter.Get(ctx, logE, &fuzzyfinder.Package{
				PackageInfo:  pkg.PackageInfo,
				RegistryName: pkg.Package.Registry,
				Version:      pkg.Package.Version,
			}, false)
			if newVersion != "" {
				newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.PackageInfo.GetName())] = newVersion
				newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.Package.Name)] = newVersion
			}
		}
		if err := c.updateFile(logE, cfgPath, newVersions); err != nil {
			return fmt.Errorf("update a package: %w", err)
		}
	}
	return nil
}

func (c *Controller) updateFile(logE *logrus.Entry, cfgFilePath string, newVersions map[string]string) error {
	b, err := afero.ReadFile(c.fs, cfgFilePath)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	updated, err := ast.UpdatePackages(logE, file, newVersions)
	if err != nil {
		return fmt.Errorf("parse a file with AST: %w", err)
	}

	if !updated {
		return nil
	}

	stat, err := c.fs.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	if err := afero.WriteFile(c.fs, cfgFilePath, []byte(file.String()), stat.Mode()); err != nil {
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}
