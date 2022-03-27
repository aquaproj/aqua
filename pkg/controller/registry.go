package controller

import (
	"context"
	"fmt"
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

func (ctrl *Controller) getGitHubContentRegistry(ctx context.Context, registry *config.Registry, registryFilePath string) (*config.RegistryContent, error) {
	b, err := ctrl.RegistryDownloader.GetGitHubContentFile(ctx, registry.RepoOwner, registry.RepoName, registry.Ref, registry.Path)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if err := os.WriteFile(registryFilePath, b, 0o600); err != nil { //nolint:gomnd
		return nil, fmt.Errorf("write the configuration file: %w", err)
	}
	registryContent := &config.RegistryContent{}
	if err := yaml.Unmarshal(b, registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration file: %w", err)
	}
	return registryContent, nil
}

// getRegistry downloads and installs the registry file.
func (ctrl *Controller) getRegistry(ctx context.Context, registry *config.Registry, registryFilePath string) (*config.RegistryContent, error) {
	switch registry.Type {
	case config.RegistryTypeGitHubContent:
		return ctrl.getGitHubContentRegistry(ctx, registry, registryFilePath)
	case config.RegistryTypeLocal:
		return nil, logerr.WithFields(errLocalRegistryNotFound, logrus.Fields{ //nolint:wrapcheck
			"local_registry_file_path": registryFilePath,
		})
	}
	return nil, errUnsupportedRegistryType
}
