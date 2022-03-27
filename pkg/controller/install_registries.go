package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

func (ctrl *Controller) installRegistries(ctx context.Context, cfg *config.Config, cfgFilePath string) (map[string]*config.RegistryContent, error) {
	var wg sync.WaitGroup
	wg.Add(len(cfg.Registries))
	var flagMutex sync.Mutex
	var registriesMutex sync.Mutex
	var failed bool
	maxInstallChan := make(chan struct{}, getMaxParallelism())
	registryContents := make(map[string]*config.RegistryContent, len(cfg.Registries)+1)

	for _, registry := range cfg.Registries {
		go func(registry *config.Registry) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			registryContent, err := ctrl.installRegistry(ctx, registry, cfgFilePath)
			if err != nil {
				<-maxInstallChan
				logerr.WithError(ctrl.logE(), err).WithFields(logrus.Fields{
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
		if err := validateRegistryContent(registryContent); err != nil {
			return nil, logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
				"registry_name": registryName,
			})
		}
	}

	return registryContents, nil
}

// installRegistry installs and reads the registry file and returns the registry content.
// If the registry file already exists, the installation is skipped.
func (ctrl *Controller) installRegistry(ctx context.Context, registry *config.Registry, cfgFilePath string) (*config.RegistryContent, error) {
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
	registryContent := &config.RegistryContent{}
	if err := yaml.NewDecoder(f).Decode(registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration: %w", err)
	}
	return registryContent, nil
}
