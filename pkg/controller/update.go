package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) Update(ctx context.Context, param *Param) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	for _, cfgFilePath := range ctrl.getConfigFilePaths(wd, param.ConfigFilePath) {
	}
	return nil
}

func (ctrl *Controller) update(ctx context.Context, rootBin, cfgFilePath string, param *Param) error {
	cfg := &Config{}
	if cfgFilePath == "" {
		return errConfigFileNotFound
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

	path, err := yaml.PathString("$.packages")
	if err != nil {
		return fmt.Errorf("build a YAML Path: %w", err)
	}

	file, err := yaml.ParseFile(cfgFilePath, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as AST: %w", err)
	}

	node, err := path.FilterFile(file)
	if err != nil {
		return fmt.Errorf("filter a node with YAML Path: %w", err)
	}
	pkgNodes, ok := node.(*ast.SequenceNode)
	if !ok {
		return errors.New("the type of packages must be array")
	}

	if len(cfg.Packages) != len(pkgNodes.Values) {
		return errors.New("len(cfg.Packages) != len(pkgNodes.Values)")
	}

	return ctrl.updatePackages(ctx, cfg, registryContents, pkgNodes)
}

func (ctrl *Controller) getFilteredLatestTag(ctx context.Context, pkgInfo *PackageInfo) string {
	if pkgInfo.VersionFilter != nil {
		return ctrl.listAndGetTagName(ctx, pkgInfo)
	}
	release, _, err := ctrl.GitHubRepositoryService.GetLatestRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName)
	if err != nil {
		logerr.WithError(ctrl.logE(), err).WithFields(logrus.Fields{
			"repo_owner": pkgInfo.RepoOwner,
			"repo_name":  pkgInfo.RepoName,
		}).Warn("get the latest release")
		return ""
	}
	return release.GetTagName()
}

func (ctrl *Controller) updatePackages(ctx context.Context, cfg *Config, registries map[string]*RegistryContent, pkgNodes *ast.SequenceNode) error {
	var failed bool
	for i, pkg := range cfg.Packages {
		logE := ctrl.logE().WithFields(logrus.Fields{
			"package_name":    pkg.Name,
			"package_version": pkg.Version,
			"registry":        pkg.Registry,
		})
		if registry, ok := cfg.Registries[pkg.Registry]; ok {
			if registry.Ref != "" {
				logE = logE.WithField("registry_ref", registry.Ref)
			}
		}
		pkgInfo, err := getPkgInfoFromRegistries(registries, pkg)
		if err != nil {
			logerr.WithError(logE, err).Error("update the package")
			failed = true
			continue
		}
		if !pkgInfo.HasRepo() {
			continue
		}
		tag := ctrl.getFilteredLatestTag(ctx, pkgInfo)
		if tag == "" || tag == pkg.Version {
			continue
		}
		node := pkgNodes.Values[i]
		pkgNode, ok := node.(*ast.MappingNode)
		if !ok {
			return errors.New("package must be a map")
		}
		tagNode, err := yaml.ValueToNode(tag)
		if err != nil {
			return err
		}
		for _, valueNode := range pkgNode.Values {
			switch valueNode.Key.String() {
			case "name":
				idx := strings.Index(valueNode.Value.String(), "@")
				if idx == -1 {
					continue
				}
				nameNode, ok := valueNode.Value.(*ast.StringNode)
				if !ok {
					return errors.New("package name must be StringNode")
				}
				nameNode.Value = valueNode.Value.String()[:idx+1] + tag
			case "version":
				valueNode.Value = tagNode
			}
		}
	}

	return nil
}
