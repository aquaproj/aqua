package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"go.yaml.in/yaml/v2"
)

const registryFilePermission = 0o600

func (is *Installer) getGitHubContentRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, registryFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	ghContentFile, err := is.registryDownloader.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
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

	// The file is written to a temporary file and renamed into place, so another aqua
	// reading the registry never sees it half written.
	if err := writeFileAtomically(registryFilePath, content); err != nil {
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
