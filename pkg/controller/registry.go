package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/google/go-github/v39/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

func (ctrl *Controller) getGitHubContentFile(ctx context.Context, repoOwner, repoName, ref, path string) ([]byte, error) {
	// https://github.com/aquaproj/aqua/issues/391
	body, err := downloadFromURL(ctx, "https://raw.githubusercontent.com/"+repoOwner+"/"+repoName+"/"+ref+"/"+path, http.DefaultClient)
	if body != nil {
		defer body.Close()
	}
	if err == nil {
		b, err := io.ReadAll(body)
		if err == nil {
			return b, nil
		}
	}

	logerr.WithError(ctrl.logE(), err).WithFields(logrus.Fields{
		"repo_owner": repoOwner,
		"repo_name":  repoName,
		"ref":        ref,
		"path":       path,
	}).Debug("failed to download a content from GitHub without GitHub API. Try again with GitHub API")

	if ctrl.GitHubRepositoryService == nil {
		return nil, errGitHubTokenIsRequired
	}

	file, _, _, err := ctrl.GitHubRepositoryService.GetContents(ctx, repoOwner, repoName, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		return nil, fmt.Errorf("get the registry configuration file by Get GitHub Content API: %w", err)
	}
	if file == nil {
		return nil, errGitHubContentMustBeFile
	}
	content, err := file.GetContent()
	if err != nil {
		return nil, fmt.Errorf("get the registry configuration content: %w", err)
	}

	return []byte(content), nil
}

func (ctrl *Controller) getGitHubContentRegistry(ctx context.Context, registry *config.Registry, registryFilePath string) (*config.RegistryContent, error) {
	b, err := ctrl.getGitHubContentFile(ctx, registry.RepoOwner, registry.RepoName, registry.Ref, registry.Path)
	if err != nil {
		return nil, err
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
