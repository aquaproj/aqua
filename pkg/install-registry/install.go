package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/aquaproj/aqua/pkg/validate"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type installer struct {
	rootDir            string
	registryDownloader download.RegistryDownloader
	logger             *log.Logger
}

func (inst *installer) logE() *logrus.Entry {
	return inst.logger.LogE()
}

func (inst *installer) InstallRegistries(ctx context.Context, cfg *config.Config, cfgFilePath string) (map[string]*config.RegistryContent, error) {
	var wg sync.WaitGroup
	wg.Add(len(cfg.Registries))
	var flagMutex sync.Mutex
	var registriesMutex sync.Mutex
	var failed bool
	maxInstallChan := make(chan struct{}, util.GetMaxParallelism())
	registryContents := make(map[string]*config.RegistryContent, len(cfg.Registries)+1)

	for _, registry := range cfg.Registries {
		go func(registry *config.Registry) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			registryContent, err := inst.installRegistry(ctx, registry, cfgFilePath)
			if err != nil {
				<-maxInstallChan
				logerr.WithError(inst.logE(), err).WithFields(logrus.Fields{
					"registry_name": registry.Name,
				}).Error("install the registry")
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

	for registryName, registryContent := range registryContents {
		if err := validate.RegistryContent(registryContent); err != nil {
			return nil, logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
				"registry_name": registryName,
			})
		}
	}

	return registryContents, nil
}

// installRegistry installs and reads the registry file and returns the registry content.
// If the registry file already exists, the installation is skipped.
func (inst *installer) installRegistry(ctx context.Context, registry *config.Registry, cfgFilePath string) (*config.RegistryContent, error) {
	registryFilePath := registry.GetFilePath(inst.rootDir, cfgFilePath)
	if err := util.MkdirAll(filepath.Dir(registryFilePath)); err != nil {
		return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
	}

	if _, err := os.Stat(registryFilePath); err != nil {
		return inst.getRegistry(ctx, registry, registryFilePath)
	}

	f, err := os.Open(registryFilePath)
	if err != nil {
		return nil, fmt.Errorf("open the registry configuration file: %w", err)
	}
	defer f.Close()
	registryContent := &config.RegistryContent{}
	if err := yaml.NewDecoder(f).Decode(registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration: %w", err)
	}
	return registryContent, nil
}

// getRegistry downloads and installs the registry file.
func (inst *installer) getRegistry(ctx context.Context, registry *config.Registry, registryFilePath string) (*config.RegistryContent, error) {
	switch registry.Type {
	case config.RegistryTypeGitHubContent:
		return inst.getGitHubContentRegistry(ctx, registry, registryFilePath)
	case config.RegistryTypeLocal:
		return nil, logerr.WithFields(errLocalRegistryNotFound, logrus.Fields{ //nolint:wrapcheck
			"local_registry_file_path": registryFilePath,
		})
	}
	return nil, errUnsupportedRegistryType
}

func (inst *installer) getGitHubContentRegistry(ctx context.Context, registry *config.Registry, registryFilePath string) (*config.RegistryContent, error) {
	b, err := inst.registryDownloader.GetGitHubContentFile(ctx, registry.RepoOwner, registry.RepoName, registry.Ref, registry.Path)
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
