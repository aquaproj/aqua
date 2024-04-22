package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

var errMaxParallelismMustBeGreaterThanZero = errors.New("MaxParallelism must be greater than zero")

func (is *Installer) InstallRegistries(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error) {
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
				logE.Debug("ignore a registry because the registry name is empty")
				return
			}
			maxInstallChan <- struct{}{}
			registryContent, err := is.installRegistry(ctx, logE, registry, cfgFilePath, checksums)
			if err != nil {
				<-maxInstallChan
				logerr.WithError(logE, err).WithFields(logrus.Fields{
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

	return registryContents, nil
}

func (is *Installer) readYAMLRegistry(p string, registry *registry.Config) error {
	f, err := is.fs.Open(p)
	if err != nil {
		return fmt.Errorf("open the registry configuration file: %w", err)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(registry); err != nil {
		return fmt.Errorf("parse the registry configuration as YAML: %w", err)
	}
	return nil
}

func (is *Installer) readJSONRegistry(p string, registry *registry.Config) error {
	f, err := is.fs.Open(p)
	if err != nil {
		return fmt.Errorf("open the registry configuration file: %w", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(registry); err != nil {
		return fmt.Errorf("parse the registry configuration as JSON: %w", err)
	}
	return nil
}

func (is *Installer) readRegistry(p string, registry *registry.Config) error {
	if isJSON(p) {
		return is.readJSONRegistry(p, registry)
	}
	return is.readYAMLRegistry(p, registry)
}

// installRegistry installs and reads the registry file and returns the registry content.
// If the registry file already exists, the installation is skipped.
func (is *Installer) installRegistry(ctx context.Context, logE *logrus.Entry, regist *aqua.Registry, cfgFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	registryFilePath, err := regist.FilePath(is.param.RootDir, cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("get a registry file path: %w", err)
	}

	if regist.Type == aqua.RegistryTypeLocal {
		registryContent := &registry.Config{}
		if err := is.readRegistry(registryFilePath, registryContent); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, logerr.WithFields(errLocalRegistryNotFound, logrus.Fields{ //nolint:wrapcheck
					"local_registry_file_path": registryFilePath,
				})
			}
			return nil, err
		}
		return registryContent, nil
	}

	if !isJSON(registryFilePath) {
		return is.handleYAMLGitHubContent(ctx, logE, regist, checksums, registryFilePath)
	}

	registryContent := &registry.Config{}
	if err := is.readJSONRegistry(registryFilePath, registryContent); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err := osfile.MkdirAll(is.fs, filepath.Dir(registryFilePath)); err != nil {
			return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
		}
		return is.getRegistry(ctx, logE, regist, registryFilePath, checksums)
	}
	return registryContent, nil
}

func (is *Installer) handleYAMLGitHubContent(ctx context.Context, logE *logrus.Entry, regist *aqua.Registry, checksums *checksum.Checksums, registryFilePath string) (*registry.Config, error) {
	jsonPath := registryFilePath + jsonSuffix
	registryContent := &registry.Config{}
	if err := is.readJSONRegistry(jsonPath, registryContent); err != nil { //nolint:nestif
		if !errors.Is(err, os.ErrNotExist) {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"registry_json_path": jsonPath,
			}).Warn("failed to read a registry JSON file. Will remove and recreate the file")
			if err := is.fs.Remove(jsonPath); err != nil {
				logerr.WithError(logE, err).WithFields(logrus.Fields{
					"registry_json_path": jsonPath,
				}).Warn("failed to remove a registry JSON file")
			} else {
				logE.WithFields(logrus.Fields{
					"registry_json_path": jsonPath,
				}).Debug("remove a registry JSON file")
			}
		}
		if err := is.readYAMLRegistry(registryFilePath, registryContent); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			if err := osfile.MkdirAll(is.fs, filepath.Dir(registryFilePath)); err != nil {
				return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
			}
			registryContent, err := is.getRegistry(ctx, logE, regist, registryFilePath, checksums)
			if err != nil {
				return nil, err
			}
			return registryContent, is.createJSON(jsonPath, registryContent)
		}
		return registryContent, is.createJSON(jsonPath, registryContent)
	}
	return registryContent, nil
}

func (is *Installer) createJSON(jsonPath string, content *registry.Config) error {
	jsonFile, err := is.fs.Create(jsonPath)
	if err != nil {
		return fmt.Errorf("create a file to convert registry YAML to JSON: %w", err)
	}
	defer jsonFile.Close()
	if err := json.NewEncoder(jsonFile).Encode(content); err != nil {
		return fmt.Errorf("encode a registry as JSON: %w", err)
	}
	return nil
}

// getRegistry downloads and installs the registry file.
func (is *Installer) getRegistry(ctx context.Context, logE *logrus.Entry, registry *aqua.Registry, registryFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	// TODO checksum verification
	// TODO download checksum file
	if registry.Type == aqua.RegistryTypeGitHubContent {
		return is.getGitHubContentRegistry(ctx, logE, registry, registryFilePath, checksums)
	}
	return nil, errUnsupportedRegistryType
}

const registryFilePermission = 0o600

func (is *Installer) getGitHubContentRegistry(ctx context.Context, logE *logrus.Entry, regist *aqua.Registry, registryFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	ghContentFile, err := is.registryDownloader.DownloadGitHubContentFile(ctx, logE, &domain.GitHubContentFileParam{
		RepoOwner: regist.RepoOwner,
		RepoName:  regist.RepoName,
		Ref:       regist.Ref,
		Path:      regist.Path,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	defer ghContentFile.Close()

	content, err := ghContentFile.Byte()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if checksums != nil {
		if err := checksum.CheckRegistry(regist, checksums, content); err != nil {
			return nil, fmt.Errorf("check a registry's checksum: %w", err)
		}
	}

	file, err := is.fs.Create(registryFilePath)
	if err != nil {
		return nil, fmt.Errorf("create a registry file: %w", err)
	}
	defer file.Close()

	if err := afero.WriteFile(is.fs, registryFilePath, content, registryFilePermission); err != nil {
		return nil, fmt.Errorf("write the configuration file: %w", err)
	}
	registryContent := &registry.Config{}
	if isJSON(registryFilePath) {
		if err := json.Unmarshal(content, registryContent); err != nil {
			return nil, fmt.Errorf("parse the registry configuration file as JSON: %w", err)
		}
		return registryContent, nil
	}
	if err := yaml.Unmarshal(content, registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration file as YAML: %w", err)
	}
	return registryContent, nil
}

func isJSON(p string) bool {
	return strings.HasSuffix(p, jsonSuffix)
}
