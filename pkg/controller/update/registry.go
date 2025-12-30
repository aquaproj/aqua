package update

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/controller/update/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/spf13/afero"
)

func (c *Controller) newRegistryVersion(ctx context.Context, logger *slog.Logger, rgst *aqua.Registry) (string, error) {
	if rgst.Type == "local" {
		return "", nil
	}

	logger.Debug("getting the latest release of a registry")
	release, _, err := c.gh.GetLatestRelease(ctx, rgst.RepoOwner, rgst.RepoName)
	if err != nil {
		return "", fmt.Errorf("get the latest release by GitHub API: %w", err)
	}
	// TODO Get the latest tag if the latest release can't be got.
	return release.GetTagName(), nil
}

func (c *Controller) updateRegistries(ctx context.Context, logger *slog.Logger, cfgFilePath string, cfg *aqua.Config) error { //nolint:cyclop
	newVersions := map[string]string{}
	for _, rgst := range cfg.Registries {
		logger := logger.With("registry_name", rgst.Name)
		if commitHashPattern.MatchString(rgst.Ref) {
			logger.Debug("skip a registry whose version is a commit hash")
			continue
		}
		newVersion, err := c.newRegistryVersion(ctx, logger, rgst)
		if err != nil {
			return err
		}
		if newVersion == "" {
			continue
		}
		newVersions[rgst.Name] = newVersion
	}

	b, err := afero.ReadFile(c.fs, cfgFilePath)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	// TODO consider how to update commit hashes
	updated, err := ast.UpdateRegistries(logger, file, newVersions)
	if err != nil {
		return fmt.Errorf("parse a configuration as YAML to update registries: %w", err)
	}

	if updated {
		stat, err := c.fs.Stat(cfgFilePath)
		if err != nil {
			return fmt.Errorf("get configuration file stat: %w", err)
		}
		if err := afero.WriteFile(c.fs, cfgFilePath, []byte(file.String()), stat.Mode()); err != nil {
			return fmt.Errorf("write the configuration file: %w", err)
		}
	}
	return nil
}
