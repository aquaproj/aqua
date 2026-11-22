package registry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var errInstallLockedOnlyGitHubContent = errors.New("only a github_content registry has a checksum in the lock file")

// InstallLockedRegistry returns the content of a github_content registry at the commit a
// lock file recorded, reading it from CommitFilePath for that commit, or downloading it
// from the commit and checking it against the recorded checksum when it isn't there yet.
//
// The file is downloaded from the commit rather than the ref, so it is what the lock file
// recorded even after the ref has moved on. A file at that path isn't checked again: only
// content of that commit checked against the checksum is ever written there, and it is
// written atomically. So, unlike the file at FilePath, it can't have been replaced by
// what the ref points at now.
func (is *Installer) InstallLockedRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, commit string, chk *checksum.Checksum) (*registry.Config, error) {
	if err := regist.Validate(); err != nil {
		return nil, fmt.Errorf("validate the registry: %w", err)
	}
	if regist.Type != aqua.RegistryTypeGitHubContent {
		return nil, errInstallLockedOnlyGitHubContent
	}
	registryFilePath, err := regist.CommitFilePath(is.param.RootDir, commit)
	if err != nil {
		return nil, fmt.Errorf("get a registry file path: %w", err)
	}

	if content, err := is.readLockedRegistry(logger, registryFilePath); err != nil || content != nil {
		return content, err
	}

	content, err := is.downloadGitHubContent(ctx, logger, regist, commit)
	if err != nil {
		return nil, err
	}
	id := checksum.RegistryID(regist)
	chks := checksum.New()
	chks.Set(id, &checksum.Checksum{ID: id, Checksum: chk.Checksum, Algorithm: chk.Algorithm})
	if err := checksum.CheckRegistry(regist, chks, content); err != nil {
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

// readLockedRegistry reads a registry kept at its CommitFilePath, the JSON made from a
// YAML registry first, as InstallRegistry does. It returns nil without an error when
// nothing is there yet.
func (is *Installer) readLockedRegistry(logger *slog.Logger, registryFilePath string) (*registry.Config, error) {
	registryContent := &registry.Config{}
	if isJSON(registryFilePath) {
		if err := is.readJSONRegistry(registryFilePath, registryContent); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, nil //nolint:nilnil
			}
			return nil, err
		}
		return registryContent, nil
	}

	jsonPath := registryFilePath + jsonSuffix
	err := is.readJSONRegistry(jsonPath, registryContent)
	if err == nil {
		return registryContent, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		// Writes are atomic, so a JSON that can't be read was damaged on disk. The YAML
		// beside it is read instead, as InstallRegistry does.
		slogerr.WithError(logger, err).Warn("failed to read a registry JSON file. Will read the YAML instead", "registry_json_path", jsonPath)
	}
	registryContent = &registry.Config{}
	if err := is.readYAMLRegistry(registryFilePath, registryContent); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil //nolint:nilnil
		}
		return nil, err
	}
	return registryContent, nil
}
