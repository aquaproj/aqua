package registry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var errMaxParallelismMustBeGreaterThanZero = errors.New("MaxParallelism must be greater than zero")

func (is *Installer) InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error) {
	var wg sync.WaitGroup
	var flagMutex sync.Mutex
	var registriesMutex sync.Mutex
	var failed bool
	if is.param.MaxParallelism <= 0 {
		return nil, errMaxParallelismMustBeGreaterThanZero
	}
	maxInstallChan := make(chan struct{}, is.param.MaxParallelism)
	registryContents := make(map[string]*registry.Config, len(cfg.Registries)+1)

	for _, registry := range cfg.Registries {
		if registry == nil {
			continue
		}
		wg.Add(1)
		go func(registry *aqua.Registry) {
			defer wg.Done()
			if registry.Name == "" {
				logger.Debug("ignore a registry because the registry name is empty")
				return
			}
			maxInstallChan <- struct{}{}
			registryContent, err := is.InstallRegistry(ctx, logger, registry, cfgFilePath, checksums)
			if err != nil {
				<-maxInstallChan
				logger.Error("install the registry",
					slog.Any("error", err),
					slog.String("registry_name", registry.Name))
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

// InstallRegistry installs and reads the registry file and returns the registry content.
// If the registry file already exists, the installation is skipped.
func (is *Installer) InstallRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, cfgFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	if err := regist.Validate(); err != nil {
		return nil, fmt.Errorf("validate the registry: %w", err)
	}
	registryFilePath, err := regist.FilePath(is.param.RootDir, cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("get a registry file path: %w", err)
	}

	if regist.Type == aqua.RegistryTypeLocal {
		registryContent := &registry.Config{}
		if err := is.readRegistry(registryFilePath, registryContent); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, slogerr.With(errLocalRegistryNotFound, //nolint:wrapcheck
					slog.String("local_registry_file_path", registryFilePath))
			}
			return nil, err
		}
		return registryContent, nil
	}

	if !isJSON(registryFilePath) {
		return is.handleYAMLGitHubContent(ctx, logger, regist, checksums, registryFilePath)
	}

	registryContent := &registry.Config{}
	if err := is.readJSONRegistry(registryFilePath, registryContent); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err := osfile.MkdirAll(is.fs, filepath.Dir(registryFilePath)); err != nil {
			return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
		}
		return is.getRegistry(ctx, logger, regist, registryFilePath, checksums)
	}
	return registryContent, nil
}

// getRegistry downloads and installs the registry file.
func (is *Installer) getRegistry(ctx context.Context, logger *slog.Logger, registry *aqua.Registry, registryFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	// TODO checksum verification
	// TODO download checksum file
	if registry.Type == aqua.RegistryTypeGitHubContent {
		return is.getGitHubContentRegistry(ctx, logger, registry, registryFilePath, checksums)
	}
	return nil, errUnsupportedRegistryType
}
