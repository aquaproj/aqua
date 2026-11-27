package update

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	errUpdatePackages   = errors.New("some packages couldn't be updated")
	errRegistryNotFound = errors.New("the registry isn't found")
	errPkgNotFound      = errors.New("the package isn't found in the registry")
)

// updatePackagesLocked updates the packages of a project with a lock file.
//
// The versions are chosen from what install reads a package from, rather than from every
// registry at its ref: aqua-registry-g2 for the standard registry, and the registry file
// at the commit the lock file recorded for any other. A version chosen that way is one
// "aqua lock update" can lock, and a registry no package to update comes from isn't read,
// so one that can't be read fails only its own packages.
//
// Every package is updated whatever environment this is: aqua.yaml is shared by all of
// them, and leaving a package behind because it can't run here would leave it behind
// everywhere.
//
// A package that can't be looked at is reported and the rest go on. The packages that
// were updated are written, and the run fails at the end.
func (c *Controller) updatePackagesLocked(ctx context.Context, logger *slog.Logger, param *config.Param, cfgFilePath string, rgsts *lockresolve.Registries) error {
	var selected map[string]struct{}
	if param.Insert {
		pkgs, err := c.selectPackages(logger, cfgFilePath)
		if err != nil {
			return err
		}
		if pkgs == nil {
			return nil
		}
		selected = pkgs
	}
	cfg := &aqua.Config{}
	cfgs, err := c.configReader.ReadToUpdate(cfgFilePath, cfg)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}
	cfgs[cfgFilePath] = cfg
	newVersions := map[string]string{}
	failed := false
	for cfgPath, cfg := range cfgs {
		ok, err := c.updateLockedPackagesInFile(ctx, logger, param, cfgPath, cfg, rgsts, selected, newVersions)
		if err != nil {
			return err
		}
		if !ok {
			failed = true
		}
	}
	if failed {
		return errUpdatePackages
	}
	return nil
}

// updateLockedPackagesInFile updates the packages a file declares, and says whether every
// package could be looked at.
func (c *Controller) updateLockedPackagesInFile(ctx context.Context, logger *slog.Logger, param *config.Param, cfgFilePath string, cfg *aqua.Config, rgsts *lockresolve.Registries, selected map[string]struct{}, newVersions map[string]string) (bool, error) { //nolint:cyclop
	ok := true
	for _, pkg := range cfg.Packages {
		if pkg.Name == "" {
			logger.Error("ignore a package because the package name is empty")
			ok = false
			continue
		}
		if pkg.Version == "" {
			logger.Error("ignore a package because the package version is empty", "package_name", pkg.Name)
			ok = false
			continue
		}
		if len(param.Args) == 0 && !param.Insert && !pkg.Update.GetEnabled() {
			continue
		}
		logger := logger.With(
			"package_name", pkg.Name,
			"package_version", pkg.Version,
			"registry", pkg.Registry,
		)
		if !aqua.FilterPackageByTag(pkg, param.Tags, param.ExcludedTags) {
			logger.Debug("skip updating the package because package tags are unmatched")
			continue
		}
		if !isSelected(selected, pkg) {
			continue
		}
		pkgInfo, source, err := versionSource(ctx, logger, rgsts, pkg)
		if err != nil {
			slogerr.WithError(logger, err).Error("read the definition of the package to update")
			ok = false
			continue
		}
		newVersion, err := c.fuzzyGetter.Get(ctx, logger, source, pkgInfo, pkg.Version, param.SelectVersion, param.Limit)
		if err != nil {
			slogerr.WithError(logger, err).Error("find the new version of the package")
			ok = false
			continue
		}
		if newVersion != "" {
			newVersions[fmt.Sprintf("%s,%s", pkg.Registry, pkgInfo.GetName())] = newVersion
			newVersions[fmt.Sprintf("%s,%s", pkg.Registry, pkg.Name)] = newVersion
		}
	}
	if len(newVersions) == 0 {
		return ok, nil
	}
	if err := c.updateFile(logger, cfgFilePath, newVersions); err != nil {
		return ok, fmt.Errorf("update a package: %w", err)
	}
	return ok, nil
}

// versionSource is the definition the new version of a package is looked for with, and
// where the versions are chosen from, as a project with a lock file reads them.
//
// The registry is looked up in the configuration the registries were read from, the
// root one. A file imported by an imported file has no registries of its own.
//
// A package of the standard registry needs only its name: aqua-registry-g2 lists its
// versions, already filtered as its definition says. A package of any other registry is
// looked for upstream with its whole definition -- version_prefix, version_filter,
// version_source and the rest decide which tags are its versions -- as the version it
// is at now has it. Overrides for an operating system or an architecture aren't applied,
// so that the version chosen doesn't depend on where aqua update runs.
func versionSource(ctx context.Context, logger *slog.Logger, rgsts *lockresolve.Registries, pkg *aqua.Package) (*registry.PackageInfo, versiongetter.Source, error) {
	rgst, ok := rgsts.Config().Registries[pkg.Registry]
	if !ok || rgst == nil {
		return nil, versiongetter.SourceUpstream, errRegistryNotFound
	}
	if lockresolve.IsStandard(rgst) {
		return &registry.PackageInfo{Name: pkg.Name}, versiongetter.SourceG2, nil
	}
	content, err := rgsts.Get(ctx, logger, pkg.Registry)
	if err != nil {
		return nil, versiongetter.SourceUpstream, err //nolint:wrapcheck
	}
	pkgInfo := content.Package(logger, pkg.Name)
	if pkgInfo == nil {
		return nil, versiongetter.SourceUpstream, errPkgNotFound
	}
	pkgInfo, err = pkgInfo.SetVersion(logger, pkg.Version)
	if err != nil {
		return nil, versiongetter.SourceUpstream, fmt.Errorf("evaluate version constraints: %w", err)
	}
	p := &config.Package{Package: pkg, PackageInfo: pkgInfo, Registry: rgst}
	if err := p.ApplyVars(); err != nil {
		return nil, versiongetter.SourceUpstream, fmt.Errorf("apply the package variable: %w", err)
	}
	return p.PackageInfo, versiongetter.SourceUpstream, nil
}

// isSelected says whether the package was chosen with -i. Every package is when nothing
// was chosen.
func isSelected(selected map[string]struct{}, pkg *aqua.Package) bool {
	if len(selected) == 0 {
		return true
	}
	_, ok := selected[selectionItem(pkg)]
	return ok
}

// selectionItem is how a package is shown when the packages to update are chosen.
func selectionItem(pkg *aqua.Package) string {
	if pkg.Registry != "standard" {
		return fmt.Sprintf("%s,%s@%s", pkg.Registry, pkg.Name, pkg.Version)
	}
	return fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
}
