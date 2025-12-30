package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (is *Installer) handleYAMLGitHubContent(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, checksums *checksum.Checksums, registryFilePath string) (*registry.Config, error) {
	jsonPath := registryFilePath + jsonSuffix
	registryContent := &registry.Config{}
	if err := is.readJSONRegistry(jsonPath, registryContent); err != nil { //nolint:nestif
		if !errors.Is(err, os.ErrNotExist) {
			slogerr.WithError(logger, err).With("registry_json_path", jsonPath).Warn("failed to read a registry JSON file. Will remove and recreate the file")
			if err := is.fs.Remove(jsonPath); err != nil {
				slogerr.WithError(logger, err).With("registry_json_path", jsonPath).Warn("failed to remove a registry JSON file")
			} else {
				logger.Debug("remove a registry JSON file",
					"registry_json_path", jsonPath)
			}
		}
		if err := is.readYAMLRegistry(registryFilePath, registryContent); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			if err := osfile.MkdirAll(is.fs, filepath.Dir(registryFilePath)); err != nil {
				return nil, fmt.Errorf("create the parent directory of the configuration file: %w", err)
			}
			registryContent, err := is.getRegistry(ctx, logger, regist, registryFilePath, checksums)
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

func isJSON(p string) bool {
	return strings.HasSuffix(p, jsonSuffix)
}
