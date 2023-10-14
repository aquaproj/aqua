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

func (c *Controller) updatePackages(ctx context.Context, logE *logrus.Entry, param *config.Param, cfgFilePath string, rgstCfgs map[string]*registry.Config) error { //nolint:cyclop
	newVersions := map[string]string{}
	for _, arg := range param.Args {
		findResult, err := c.which.Which(ctx, logE, param, arg)
		if err != nil {
			return fmt.Errorf("find a command: %w", err)
		}
		pkg := findResult.Package
		if newVersion := c.getPackageNewVersion(ctx, logE, param, nil, pkg); newVersion != "" {
			newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.PackageInfo.GetName())] = newVersion
			newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.Package.Name)] = newVersion
		}
		filePath := cfgFilePath
		if pkg.Package.FilePath != "" {
			filePath = pkg.Package.FilePath
		}
		if err := c.updateFile(logE, filePath, newVersions); err != nil {
			return fmt.Errorf("update a package: %w", err)
		}
	}
	if len(param.Args) != 0 {
		return nil
	}
	updatedPkgs := map[string]struct{}{}
	if param.Insert {
		pkgs, err := c.selectPackages(logE, cfgFilePath)
		if err != nil {
			return err
		}
		if pkgs == nil {
			return nil
		}
		updatedPkgs = pkgs
	}
	cfg := &aqua.Config{}
	cfgs, err := c.configReader.ReadToUpdate(cfgFilePath, cfg)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}
	cfgs[cfgFilePath] = cfg
	for cfgPath, cfg := range cfgs {
		if err := c.updatePackagesInFile(ctx, logE, param, cfgPath, cfg, rgstCfgs, updatedPkgs, newVersions); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) updatePackagesInFile(ctx context.Context, logE *logrus.Entry, param *config.Param, cfgFilePath string, cfg *aqua.Config, rgstCfgs map[string]*registry.Config, updatedPkgs map[string]struct{}, newVersions map[string]string) error {
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
		if newVersion := c.getPackageNewVersion(ctx, logE, param, updatedPkgs, pkg); newVersion != "" {
			newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.PackageInfo.GetName())] = newVersion
			newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.Package.Name)] = newVersion
		}
	}
	if err := c.updateFile(logE, cfgFilePath, newVersions); err != nil {
		return fmt.Errorf("update a package: %w", err)
	}
	return nil
}

func (c *Controller) getPackageNewVersion(ctx context.Context, logE *logrus.Entry, param *config.Param, updatedPkgs map[string]struct{}, pkg *config.Package) string {
	if len(updatedPkgs) != 0 {
		var item string
		if pkg.Package.Registry != "standard" {
			item = fmt.Sprintf("%s,%s@%s", pkg.Package.Registry, pkg.Package.Name, pkg.Package.Version)
		} else {
			item = fmt.Sprintf("%s@%s", pkg.Package.Name, pkg.Package.Version)
		}
		if _, ok := updatedPkgs[item]; !ok {
			return ""
		}
	}
	return c.fuzzyGetter.Get(ctx, logE, &fuzzyfinder.Package{
		PackageInfo:  pkg.PackageInfo,
		RegistryName: pkg.Package.Registry,
		Version:      pkg.Package.Version,
	}, param.SelectVersion)
}

func (c *Controller) selectPackages(logE *logrus.Entry, cfgFilePath string) (map[string]struct{}, error) {
	updatedPkgs := map[string]struct{}{}
	cfg := &aqua.Config{}
	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return nil, fmt.Errorf("read a configuration file: %w", err)
	}
	items := make([]*fuzzyfinder.Item, 0, len(cfg.Packages))
	for _, pkg := range cfg.Packages {
		if commitHashPattern.MatchString(pkg.Version) {
			logE.WithFields(logrus.Fields{
				"registry_name":   pkg.Registry,
				"package_name":    pkg.Name,
				"package_version": pkg.Version,
			}).Debug("skip a package whose version is a commit hash")
			continue
		}
		var item string
		if pkg.Registry != "standard" {
			item = fmt.Sprintf("%s,%s@%s", pkg.Registry, pkg.Name, pkg.Version)
		} else {
			item = fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
		}
		items = append(items, &fuzzyfinder.Item{
			Item: item,
		})
	}
	idxs, err := c.fuzzyFinder.FindMulti(items, false)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil, nil //nolint:nilnil
		}
		return nil, fmt.Errorf("select updated packages with fuzzy finder: %w", err)
	}
	for _, idx := range idxs {
		updatedPkgs[items[idx].Item] = struct{}{}
	}
	return updatedPkgs, nil
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
