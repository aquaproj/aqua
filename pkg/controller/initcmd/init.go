package initcmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const configTemplate = `---
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/aqua-yaml.json
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
# checksum:
#   enabled: true
#   require_checksum: true
#   supported_envs:
#   - all
registries:
- type: standard
  ref: %%STANDARD_REGISTRY_VERSION%%  # renovate: depName=aquaproj/aqua-registry
packages:
`

const importDirConfigTemplate = `---
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/aqua-yaml.json
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
# checksum:
#   enabled: true
#   require_checksum: true
#   supported_envs:
#   - all
registries:
- type: standard
  ref: %%STANDARD_REGISTRY_VERSION%%  # renovate: depName=aquaproj/aqua-registry
import_dir: %%IMPORT_DIR%%
`

type Controller struct {
	github RepositoriesService
	fs     afero.Fs
}

func New(gh RepositoriesService, fs afero.Fs) *Controller {
	return &Controller{
		github: gh,
		fs:     fs,
	}
}

type Param struct {
	ImportDir string
	IsDir     bool
}

func (c *Controller) Init(ctx context.Context, logE *logrus.Entry, cfgFilePath string, param *Param) error {
	cfgFilePath = c.cfgFilePath(cfgFilePath, param)

	for _, name := range append(finder.DuplicateFilePaths(cfgFilePath), cfgFilePath) {
		if _, err := c.fs.Stat(name); err == nil {
			// configuration file already exists, then do nothing.
			logE.WithFields(logrus.Fields{
				"configuration_file_path": name,
			}).Info("configuration file already exists")
			return nil
		}
	}

	if param.IsDir {
		if err := osfile.MkdirAll(c.fs, "aqua"); err != nil {
			return err //nolint:wrapcheck
		}
	}

	registryVersion := "v4.443.0" // renovate: depName=aquaproj/aqua-registry
	release, _, err := c.github.GetLatestRelease(ctx, "aquaproj", "aqua-registry")
	if err != nil {
		logerr.WithError(logE, err).WithFields(logrus.Fields{
			"repo_owner": "aquaproj",
			"repo_name":  "aqua-registry",
		}).Warn("get the latest release")
	} else {
		if release == nil {
			logE.WithFields(logrus.Fields{
				"repo_owner": "aquaproj",
				"repo_name":  "aqua-registry",
			}).Warn("failed to get the latest release")
		} else {
			registryVersion = release.GetTagName()
		}
	}
	var cfgStr string
	if param.ImportDir == "" {
		cfgStr = strings.Replace(configTemplate, "%%STANDARD_REGISTRY_VERSION%%", registryVersion, 1)
	} else {
		cfgStr = strings.Replace(
			strings.Replace(importDirConfigTemplate, "%%STANDARD_REGISTRY_VERSION%%", registryVersion, 1),
			"%%IMPORT_DIR%%", param.ImportDir, 1)
	}
	if err := afero.WriteFile(c.fs, cfgFilePath, []byte(cfgStr), osfile.FilePermission); err != nil {
		return fmt.Errorf("write a configuration file: %w", logerr.WithFields(err, logrus.Fields{
			"configuration_file_path": cfgFilePath,
		}))
	}
	return nil
}

func (c *Controller) cfgFilePath(cfgFilePath string, param *Param) string {
	if cfgFilePath != "" {
		return cfgFilePath
	}
	if !param.IsDir {
		return "aqua.yaml"
	}
	return filepath.Join("aqua", "aqua.yaml")
}
