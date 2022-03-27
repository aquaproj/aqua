package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/v39/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type Registry struct {
	Name      string `validate:"required"`
	Type      string `validate:"required"`
	RepoOwner string `yaml:"repo_owner"`
	RepoName  string `yaml:"repo_name"`
	Ref       string
	Path      string `validate:"required"`
}

const (
	registryTypeGitHubContent = "github_content"
	registryTypeLocal         = "local"
	registryTypeStandard      = "standard"
)

func (registry *Registry) validate() error {
	switch registry.Type {
	case registryTypeLocal:
		return registry.validateLocal()
	case registryTypeGitHubContent:
		return registry.validateGitHubContent()
	default:
		return logerr.WithFields(errInvalidRegistryType, logrus.Fields{ //nolint:wrapcheck
			"registry_type": registry.Type,
		})
	}
}

func (registry *Registry) validateLocal() error {
	if registry.Path == "" {
		return errPathIsRequired
	}
	return nil
}

func (registry *Registry) validateGitHubContent() error {
	if registry.RepoOwner == "" {
		return errRepoOwnerIsRequired
	}
	if registry.RepoName == "" {
		return errRepoNameIsRequired
	}
	if registry.Ref == "" {
		return errRefIsRequired
	}
	return nil
}

func (registry *Registry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Registry
	a := alias(*registry)
	if err := unmarshal(&a); err != nil {
		return err
	}
	if a.Type == registryTypeStandard {
		if a.Name == "" {
			a.Name = registryTypeStandard
		}
		a.Type = registryTypeGitHubContent
		a.RepoOwner = "aquaproj"
		a.RepoName = "aqua-registry"
		a.Path = "registry.yaml"
	}
	*registry = Registry(a)
	return nil
}

func (registry *Registry) GetFilePath(rootDir, cfgFilePath string) string {
	switch registry.Type {
	case registryTypeLocal:
		if filepath.IsAbs(registry.Path) {
			return registry.Path
		}
		return filepath.Join(filepath.Dir(cfgFilePath), registry.Path)
	case registryTypeGitHubContent:
		return filepath.Join(rootDir, "registries", registry.Type, "github.com", registry.RepoOwner, registry.RepoName, registry.Ref, registry.Path)
	}
	return ""
}

type RegistryContent struct {
	PackageInfos PackageInfos `yaml:"packages" validate:"dive"`
}

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

func (ctrl *Controller) getGitHubContentRegistry(ctx context.Context, registry *Registry, registryFilePath string) (*RegistryContent, error) {
	b, err := ctrl.getGitHubContentFile(ctx, registry.RepoOwner, registry.RepoName, registry.Ref, registry.Path)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(registryFilePath, b, 0o600); err != nil { //nolint:gomnd
		return nil, fmt.Errorf("write the configuration file: %w", err)
	}
	registryContent := &RegistryContent{}
	if err := yaml.Unmarshal(b, registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration file: %w", err)
	}
	return registryContent, nil
}

// getRegistry downloads and installs the registry file.
func (ctrl *Controller) getRegistry(ctx context.Context, registry *Registry, registryFilePath string) (*RegistryContent, error) {
	switch registry.Type {
	case registryTypeGitHubContent:
		return ctrl.getGitHubContentRegistry(ctx, registry, registryFilePath)
	case registryTypeLocal:
		return nil, logerr.WithFields(errLocalRegistryNotFound, logrus.Fields{ //nolint:wrapcheck
			"local_registry_file_path": registryFilePath,
		})
	}
	return nil, errUnsupportedRegistryType
}
