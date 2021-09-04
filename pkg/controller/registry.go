package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/go-github/v38/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/aqua/pkg/log"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type Registry interface {
	GetName() string
	GetType() string
	GetFilePath(rootDir, cfgFilePath string) string
}

type mergedRegistry struct {
	Name      string `validate:"required"`
	Type      string `validate:"required"`
	RepoOwner string `yaml:"repo_owner"`
	RepoName  string `yaml:"repo_name"`
	Ref       string `validate:"required"`
	Path      string `validate:"required"`
}

const (
	registryTypeGitHubContent = "github_content"
	registryTypeLocal         = "local"
	registryTypeStandard      = "standard"
)

func (registry *mergedRegistry) GetRegistry() (Registry, error) {
	switch registry.Type {
	case registryTypeGitHubContent:
		return &GitHubContentRegistry{
			Name:      registry.Name,
			RepoOwner: registry.RepoOwner,
			RepoName:  registry.RepoName,
			Ref:       registry.Ref,
			Path:      registry.Path,
		}, nil
	case registryTypeLocal:
		return &LocalRegistry{
			Name: registry.Name,
			Path: registry.Path,
		}, nil
	case registryTypeStandard:
		return &GitHubContentRegistry{
			Name:      "standard",
			RepoOwner: "suzuki-shunsuke",
			RepoName:  "aqua-registry",
			Ref:       registry.Ref,
			Path:      "registry.yaml",
		}, nil
	default:
		return nil, errors.New("type is invalid")
	}
}

type GitHubContentRegistry struct {
	Name      string `validate:"required"`
	RepoOwner string `yaml:"repo_owner"`
	RepoName  string `yaml:"repo_name"`
	Ref       string `validate:"required"`
	Path      string `validate:"required"`
}

func (registry *GitHubContentRegistry) GetName() string {
	return registry.Name
}

func (registry *GitHubContentRegistry) GetType() string {
	return registryTypeGitHubContent
}

func (registry *GitHubContentRegistry) GetFilePath(rootDir, cfgFilePath string) string {
	return filepath.Join(rootDir, "registries", registry.GetType(), "github.com", registry.RepoOwner, registry.RepoName, registry.Ref, registry.Path)
}

type LocalRegistry struct {
	Name string `validate:"required"`
	Path string `validate:"required"`
}

func (registry *LocalRegistry) GetName() string {
	return registry.Name
}

func (registry *LocalRegistry) GetType() string {
	return registryTypeLocal
}

func (registry *LocalRegistry) GetFilePath(rootDir, cfgFilePath string) string {
	if filepath.IsAbs(registry.Path) {
		return registry.Path
	}
	return filepath.Join(filepath.Dir(cfgFilePath), registry.Path)
}

type RegistryContent struct {
	PackageInfos PackageInfos `yaml:"packages" validate:"dive"`
}

func (ctrl *Controller) installRegistries(ctx context.Context, cfg *Config, cfgFilePath string) (map[string]*RegistryContent, error) {
	var wg sync.WaitGroup
	wg.Add(len(cfg.Registries))
	var flagMutex sync.Mutex
	var registriesMutex sync.Mutex
	var failed bool
	maxInstallChan := make(chan struct{}, getMaxParallelism())
	registryContents := make(map[string]*RegistryContent, len(cfg.Registries)+1)
	registryContents["inline"] = cfg.InlineRegistry

	for _, registry := range cfg.Registries {
		go func(registry Registry) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			registryContent, err := ctrl.installRegistry(ctx, registry, cfgFilePath)
			if err != nil {
				<-maxInstallChan
				log.New().WithFields(logrus.Fields{
					"registry_name": registry.GetName(),
				}).WithError(err).Error("install the registry")
				flagMutex.Lock()
				failed = true
				flagMutex.Unlock()
				return
			}
			registriesMutex.Lock()
			registryContents[registry.GetName()] = registryContent
			registriesMutex.Unlock()
			<-maxInstallChan
		}(registry)
	}
	wg.Wait()
	if failed {
		return nil, errInstallFailure
	}

	return registryContents, nil
}

func (ctrl *Controller) getGitHubContentRegistry(ctx context.Context, registry Registry, registryFilePath string) (*RegistryContent, error) {
	if ctrl.GitHub == nil {
		return nil, errGitHubTokenIsRequired
	}
	r, ok := registry.(*GitHubContentRegistry)
	if !ok {
		return nil, errors.New("registry.GetType() is github_content, but registry isn't *GitHubContentRegistry")
	}

	file, _, _, err := ctrl.GitHub.Repositories.GetContents(ctx, r.RepoOwner, r.RepoName, r.Path, &github.RepositoryContentGetOptions{
		Ref: r.Ref,
	})
	if err != nil {
		return nil, fmt.Errorf("get the registry configuration file by Get GitHub Content API: %w", err)
	}
	if file == nil {
		return nil, errors.New("ref must be not a directory but a file")
	}
	content, err := file.GetContent()
	if err != nil {
		return nil, fmt.Errorf("get the registry configuration content: %w", err)
	}
	if err := os.WriteFile(registryFilePath, []byte(content), 0o600); err != nil { //nolint:gomnd
		return nil, fmt.Errorf("write the configuration file: %w", err)
	}
	registryContent := &RegistryContent{}
	if err := yaml.Unmarshal([]byte(content), registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration file: %w", err)
	}
	return registryContent, nil
}

func (ctrl *Controller) getRegistry(ctx context.Context, registry Registry, registryFilePath string) (*RegistryContent, error) {
	// file doesn't exist
	// download and install file
	switch registry.GetType() {
	case registryTypeGitHubContent:
		return ctrl.getGitHubContentRegistry(ctx, registry, registryFilePath)
	case registryTypeLocal:
		return nil, logerr.WithFields(errors.New("local registry isn't found"), logrus.Fields{ //nolint:wrapcheck
			"local_registry_file_path": registryFilePath,
		})
	}
	return nil, errors.New("unsupported registry type")
}

func (ctrl *Controller) installRegistry(ctx context.Context, registry Registry, cfgFilePath string) (*RegistryContent, error) {
	registryFilePath := registry.GetFilePath(ctrl.RootDir, cfgFilePath)
	if err := mkdirAll(filepath.Dir(registryFilePath)); err != nil {
		return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
	}

	if _, err := os.Stat(registryFilePath); err != nil {
		return ctrl.getRegistry(ctx, registry, registryFilePath)
	}

	f, err := os.Open(registryFilePath)
	if err != nil {
		return nil, fmt.Errorf("open the registry configuration file: %w", err)
	}
	defer f.Close()
	registryContent := &RegistryContent{}
	decoder := yaml.NewDecoder(f)
	decoder.SetStrict(true)
	if err := decoder.Decode(registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration: %w", err)
	}
	return registryContent, nil
}
