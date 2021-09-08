package controller

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type FindingPackage struct {
	PackageInfo  PackageInfo
	RegistryName string
}

func (ctrl *Controller) Generate(ctx context.Context, param *Param) error { //nolint:cyclop
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	cfgFilePath := ctrl.getConfigFilePath(wd, param.ConfigFilePath)
	if cfgFilePath == "" {
		return errConfigFileNotFound
	}
	cfg := &Config{}
	if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
		return err
	}
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}
	registryContents, err := ctrl.installRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return err
	}
	var pkgs []*FindingPackage
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			pkgs = append(pkgs, &FindingPackage{
				PackageInfo:  pkg,
				RegistryName: registryName,
			})
		}
	}
	idx, err := fuzzyfinder.Find(pkgs, func(i int) string {
		pkg := pkgs[i]
		return fmt.Sprintf("%s (%s)", pkg.PackageInfo.GetName(), pkg.RegistryName)
	},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			pkg := pkgs[i]
			return fmt.Sprintf("%s\n\n%s\n%s",
				pkg.PackageInfo.GetName(),
				pkg.PackageInfo.GetLink(),
				pkg.PackageInfo.GetDescription())
		}))
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return fmt.Errorf("find the package: %w", err)
	}
	pkg := pkgs[idx]
	outputPkg, err := ctrl.getOutputtedPkg(ctx, pkg)
	if err != nil {
		return err
	}
	if err := yaml.NewEncoder(ctrl.Stdout).Encode([]interface{}{outputPkg}); err != nil {
		return fmt.Errorf("output generated package configuration: %w", err)
	}
	return nil
}

func (ctrl *Controller) getOutputtedPkg(ctx context.Context, pkg *FindingPackage) (*Package, error) {
	outputPkg := &Package{
		Name:     pkg.PackageInfo.GetName(),
		Registry: pkg.RegistryName,
	}
	if pkg.PackageInfo.GetType() != pkgInfoTypeGitHubRelease {
		return outputPkg, nil
	}
	if ctrl.GitHub == nil {
		return outputPkg, nil
	}
	p, ok := pkg.PackageInfo.(*GitHubReleasePackageInfo)
	if !ok {
		return nil, errGitHubReleaseTypeAssertion
	}
	releases, _, err := ctrl.GitHub.Repositories.ListReleases(ctx, p.RepoOwner, p.RepoName, nil)
	if err != nil {
		logrus.WithError(err).Warn("list releases")
		return outputPkg, nil
	}
	idx, err := fuzzyfinder.Find(releases, func(i int) string {
		release := releases[i]
		if release.GetPrerelease() {
			return release.GetTagName() + " (prerelease)"
		}
		return release.GetTagName()
	})
	if err != nil {
		if !errors.Is(err, fuzzyfinder.ErrAbort) {
			logrus.WithError(err).Warn("find versions")
		}
		return outputPkg, nil
	}
	outputPkg.Version = releases[idx].GetTagName()
	return outputPkg, nil
}
