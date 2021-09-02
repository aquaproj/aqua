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
	"gopkg.in/yaml.v2"
)

type Registry struct {
	Name      string `validate:"required"`
	Type      string `validate:"required"`
	RepoOwner string `yaml:"repo_owner"`
	RepoName  string `yaml:"repo_name"`
	Ref       string `validate:"required"`
	Path      string `validate:"required"`
}

type RegistryContent struct {
	PackageInfos PackageInfos `yaml:"packages"`
}

func (ctrl *Controller) installRegistries(ctx context.Context, cfg *Config) (map[string]*RegistryContent, error) {
	var wg sync.WaitGroup
	wg.Add(len(cfg.Registries))
	var flagMutex sync.Mutex
	var registriesMutex sync.Mutex
	var failed bool
	maxInstallChan := make(chan struct{}, getMaxParallelism())
	registryContents := make(map[string]*RegistryContent, len(cfg.Registries)+1)
	registryContents["inline"] = &RegistryContent{
		PackageInfos: cfg.InlineRegistry,
	}

	for _, registry := range cfg.Registries {
		go func(registry *Registry) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			registryContent, err := ctrl.installRegistry(ctx, registry)
			if err != nil {
				<-maxInstallChan
				log.New().WithFields(logrus.Fields{
					"registry_name": registry.Name,
				}).WithError(err).Error("install the registry")
				flagMutex.Lock()
				failed = true
				flagMutex.Unlock()
				return
			}
			registriesMutex.Lock()
			registryContents[registry.Name] = registryContent
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

func (ctrl *Controller) installRegistry(ctx context.Context, registry *Registry) (*RegistryContent, error) { //nolint:cyclop
	if registry.Type != "github_content" {
		return nil, fmt.Errorf("only inline or github_content registry is supported (%s)", registry.Type)
	}

	// TODO check if file exists
	registryFilePath := filepath.Join(ctrl.RootDir, "registries", registry.Type, "github.com", registry.RepoOwner, registry.RepoName, registry.Ref, registry.Path)
	if _, err := os.Stat(registryFilePath); err != nil { //nolint:nestif
		// file doesn't exist
		// download and install file
		if ctrl.GitHub == nil {
			return nil, errGitHubTokenIsRequired
		}
		file, _, _, err := ctrl.GitHub.Repositories.GetContents(ctx, registry.RepoOwner, registry.RepoName, registry.Path, &github.RepositoryContentGetOptions{
			Ref: registry.Ref,
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

		if err := mkdirAll(filepath.Dir(registryFilePath)); err != nil {
			return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
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
