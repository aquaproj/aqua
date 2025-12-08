package registry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v2"
)

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
