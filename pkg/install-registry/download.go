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

var (
	errDownloadOnlyGitHubContent = errors.New("only a github_content registry can be downloaded")
	errNoDownloadedChecksum      = errors.New("the checksum of the downloaded registry file wasn't recorded")
)

// DownloadRegistry downloads a github_content registry whatever is installed, checks it
// against checksums the way InstallRegistry does (recording a checksum the set doesn't
// have), and keeps it at LockedFilePath for its checksum.
//
// It is for recording a registry's checksum, which has to be the checksum of what the
// registry serves: a file already on disk may be stale or edited. The file is kept under
// its own checksum rather than in place of what the ref pointed at before, so a lock file
// that recorded the earlier content still reads that content, whether the run recording
// the new one succeeds or not.
func (is *Installer) DownloadRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, _ string, checksums *checksum.Checksums) (*registry.Config, error) {
	if err := regist.Validate(); err != nil {
		return nil, fmt.Errorf("validate the registry: %w", err)
	}
	if regist.Type != aqua.RegistryTypeGitHubContent {
		return nil, errDownloadOnlyGitHubContent
	}

	content, err := is.downloadGitHubContent(ctx, logger, regist)
	if err != nil {
		return nil, err
	}
	if err := checksum.CheckRegistry(regist, checksums, content); err != nil {
		return nil, fmt.Errorf("check a registry's checksum: %w", err)
	}
	chk := checksums.Get(checksum.RegistryID(regist))
	if chk == nil {
		return nil, errNoDownloadedChecksum
	}
	registryFilePath, err := regist.LockedFilePath(is.param.RootDir, chk.Algorithm, chk.Checksum)
	if err != nil {
		return nil, fmt.Errorf("get a registry file path: %w", err)
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

func (is *Installer) downloadGitHubContent(ctx context.Context, logger *slog.Logger, regist *aqua.Registry) ([]byte, error) {
	file, err := is.registryDownloader.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
		RepoOwner: regist.RepoOwner,
		RepoName:  regist.RepoName,
		Ref:       regist.Ref,
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
