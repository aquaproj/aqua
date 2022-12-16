package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type Installer struct {
	registryDownloader domain.GitHubContentFileDownloader
	param              *config.Param
	fs                 afero.Fs
	cosign             CosignVerifier
	rt                 *runtime.Runtime
}

type CosignVerifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, b []byte) error
	HasCosign() bool
}

var errMaxParallelismMustBeGreaterThanZero = errors.New("MaxParallelism must be greater than zero")

func (inst *Installer) InstallRegistries(ctx context.Context, cfg *aqua.Config, cfgFilePath string, logE *logrus.Entry) (map[string]*registry.Config, error) {
	var wg sync.WaitGroup
	var flagMutex sync.Mutex
	var registriesMutex sync.Mutex
	var failed bool
	if inst.param.MaxParallelism <= 0 {
		return nil, errMaxParallelismMustBeGreaterThanZero
	}
	maxInstallChan := make(chan struct{}, inst.param.MaxParallelism)
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
			registryContent, err := inst.installRegistry(ctx, registry, cfgFilePath, logE)
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

func (inst *Installer) readRegistry(p string, registry *registry.Config) error {
	f, err := inst.fs.Open(p)
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

const dirPermission os.FileMode = 0o775

// installRegistry installs and reads the registry file and returns the registry content.
// If the registry file already exists, the installation is skipped.
func (inst *Installer) installRegistry(ctx context.Context, regist *aqua.Registry, cfgFilePath string, logE *logrus.Entry) (*registry.Config, error) {
	registryFilePath, err := regist.GetFilePath(inst.param.RootDir, cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("get a registry file path: %w", err)
	}
	if _, err := inst.fs.Stat(registryFilePath); err == nil {
		registryContent := &registry.Config{}
		if err := inst.readRegistry(registryFilePath, registryContent); err != nil {
			return nil, err
		}
		return registryContent, nil
	}
	if err := inst.fs.MkdirAll(filepath.Dir(registryFilePath), dirPermission); err != nil {
		return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
	}
	return inst.getRegistry(ctx, regist, registryFilePath, logE)
}

// getRegistry downloads and installs the registry file.
func (inst *Installer) getRegistry(ctx context.Context, registry *aqua.Registry, registryFilePath string, logE *logrus.Entry) (*registry.Config, error) {
	switch registry.Type {
	case aqua.RegistryTypeGitHubContent:
		return inst.getGitHubContentRegistry(ctx, registry, registryFilePath, logE)
	case aqua.RegistryTypeLocal:
		return nil, logerr.WithFields(errLocalRegistryNotFound, logrus.Fields{ //nolint:wrapcheck
			"local_registry_file_path": registryFilePath,
		})
	}
	return nil, errUnsupportedRegistryType
}

const registryFilePermission = 0o600

func (inst *Installer) getGitHubContentRegistry(ctx context.Context, regist *aqua.Registry, registryFilePath string, logE *logrus.Entry) (*registry.Config, error) { //nolint:cyclop,funlen
	ghContentFile, err := inst.registryDownloader.DownloadGitHubContentFile(ctx, logE, &domain.GitHubContentFileParam{
		RepoOwner: regist.RepoOwner,
		RepoName:  regist.RepoName,
		Ref:       regist.Ref,
		Path:      regist.Path,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	var content []byte

	if ghContentFile.String != "" {
		content = []byte(ghContentFile.String)
	} else {
		defer ghContentFile.ReadCloser.Close()
		cnt, err := io.ReadAll(ghContentFile.ReadCloser)
		if err != nil {
			return nil, fmt.Errorf("read the registry configuration file: %w", err)
		}
		content = cnt
	}

	if regist.Cosign != nil && inst.cosign.HasCosign() {
		art := &template.Artifact{
			Version: regist.Ref,
			Asset:   regist.Path,
		}
		if err := inst.cosign.Verify(ctx, logE, inst.rt, &download.File{
			RepoOwner: regist.RepoOwner,
			RepoName:  regist.RepoName,
			Version:   regist.Ref,
		}, regist.Cosign, art, content); err != nil {
			return nil, fmt.Errorf("verify a registry with Cosign: %w", err)
		}
	}

	file, err := inst.fs.Create(registryFilePath)
	if err != nil {
		return nil, fmt.Errorf("create a registry file: %w", err)
	}
	defer file.Close()

	if err := afero.WriteFile(inst.fs, registryFilePath, content, registryFilePermission); err != nil {
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
