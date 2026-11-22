package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"go.yaml.in/yaml/v2"
)

var errDownloadOnlyGitHubContent = errors.New("only a github_content registry can be downloaded")

// DownloadRegistry downloads a github_content registry from a commit whatever is
// installed, checks it against checksums the way InstallRegistry does (recording a
// checksum the set doesn't have), and keeps it at CommitFilePath for the commit.
//
// It is for recording a registry's checksum, which has to be the checksum of what the
// commit holds: a file already on disk may be stale or edited. It is downloaded from the
// commit the ref was resolved to rather than the ref, which may have moved on since, or
// be served stale by a cache, and the checksum would then be recorded beside a commit it
// doesn't describe.
func (is *Installer) DownloadRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, commit string, checksums *checksum.Checksums) (*registry.Config, error) {
	if err := regist.Validate(); err != nil {
		return nil, fmt.Errorf("validate the registry: %w", err)
	}
	if regist.Type != aqua.RegistryTypeGitHubContent {
		return nil, errDownloadOnlyGitHubContent
	}

	registryFilePath, err := regist.CommitFilePath(is.param.RootDir, commit)
	if err != nil {
		return nil, fmt.Errorf("get a registry file path: %w", err)
	}

	content, err := is.downloadGitHubContent(ctx, logger, regist, commit)
	if err != nil {
		return nil, err
	}
	if err := checksum.CheckRegistry(regist, checksums, content); err != nil {
		return nil, fmt.Errorf("check a registry's checksum: %w", err)
	}

	registryContent, err := parseRegistry(registryFilePath, content)
	if err != nil {
		return nil, err
	}
	if err := writeLockedRegistry(registryFilePath, content, registryContent); err != nil {
		return nil, err
	}
	return registryContent, nil
}

func parseRegistry(registryFilePath string, content []byte) (*registry.Config, error) {
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

// downloadGitHubContent downloads a registry file at ref, which is the commit the lock file
// records rather than the registry's own ref.
func (is *Installer) downloadGitHubContent(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, ref string) ([]byte, error) {
	file, err := is.registryDownloader.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
		RepoOwner: regist.RepoOwner,
		RepoName:  regist.RepoName,
		Ref:       ref,
		Path:      regist.Path,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	defer file.Close()
	content, err := file.Byte()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return content, nil
}
