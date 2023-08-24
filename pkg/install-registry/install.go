package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type InstallerImpl struct {
	registryDownloader domain.GitHubContentFileDownloader
	param              *config.Param
	fs                 afero.Fs
	cosign             cosign.Verifier
	slsaVerifier       slsa.Verifier
	rt                 *runtime.Runtime
}

func New(param *config.Param, downloader domain.GitHubContentFileDownloader, fs afero.Fs, rt *runtime.Runtime, cos cosign.Verifier, slsaVerifier slsa.Verifier) *InstallerImpl {
	return &InstallerImpl{
		param:              param,
		registryDownloader: downloader,
		fs:                 fs,
		rt:                 rt,
		cosign:             cos,
		slsaVerifier:       slsaVerifier,
	}
}

type Installer interface {
	InstallRegistries(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}

type MockInstaller struct {
	M   map[string]*registry.Config
	Err error
}

func (m *MockInstaller) InstallRegistries(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error) {
	return m.M, m.Err
}

var errMaxParallelismMustBeGreaterThanZero = errors.New("MaxParallelism must be greater than zero")

func (is *InstallerImpl) InstallRegistries(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error) {
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

func (is *InstallerImpl) readRegistry(p string, registry *registry.Config) error {
	f, err := is.fs.Open(p)
	if err != nil {
		return fmt.Errorf("open the registry configuration file: %w", err)
	}
	defer f.Close()
	if filepath.Ext(p) == ".json" {
		if err := json.NewDecoder(f).Decode(registry); err != nil {
			return fmt.Errorf("parse the registry configuration as JSON: %w", err)
		}
		return nil
	}
	if err := yaml.NewDecoder(f).Decode(registry); err != nil {
		return fmt.Errorf("parse the registry configuration as YAML: %w", err)
	}
	return nil
}

// installRegistry installs and reads the registry file and returns the registry content.
// If the registry file already exists, the installation is skipped.
func (is *InstallerImpl) installRegistry(ctx context.Context, logE *logrus.Entry, regist *aqua.Registry, cfgFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	registryFilePath, err := regist.GetFilePath(is.param.RootDir, cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("get a registry file path: %w", err)
	}
	if _, err := is.fs.Stat(registryFilePath); err == nil {
		registryContent := &registry.Config{}
		if err := is.readRegistry(registryFilePath, registryContent); err != nil {
			return nil, err
		}
		return registryContent, nil
	}
	if regist.Type == aqua.RegistryTypeLocal {
		return nil, logerr.WithFields(errLocalRegistryNotFound, logrus.Fields{ //nolint:wrapcheck
			"local_registry_file_path": registryFilePath,
		})
	}
	if err := util.MkdirAll(is.fs, filepath.Dir(registryFilePath)); err != nil {
		return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
	}
	return is.getRegistry(ctx, logE, regist, registryFilePath, checksums)
}

// getRegistry downloads and installs the registry file.
func (is *InstallerImpl) getRegistry(ctx context.Context, logE *logrus.Entry, registry *aqua.Registry, registryFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	// TODO checksum verification
	// TODO download checksum file
	if registry.Type == aqua.RegistryTypeGitHubContent {
		return is.getGitHubContentRegistry(ctx, logE, registry, registryFilePath, checksums)
	}
	return nil, errUnsupportedRegistryType
}

const registryFilePermission = 0o600

func (is *InstallerImpl) getGitHubContentRegistry(ctx context.Context, logE *logrus.Entry, regist *aqua.Registry, registryFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
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
	if filepath.Ext(registryFilePath) == ".json" {
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
