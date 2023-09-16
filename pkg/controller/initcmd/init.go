package initcmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const configTemplate = `---
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
# checksum:
#   enabled: true
#   require_checksum: true
#   supported_envs:
#   - all
registries:
- type: standard
  ref: %%STANDARD_REGISTRY_VERSION%% # renovate: depName=aquaproj/aqua-registry
packages:
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

func (c *Controller) Init(ctx context.Context, cfgFilePath string, logE *logrus.Entry) error {
	if cfgFilePath == "" {
		cfgFilePath = "aqua.yaml"
	}
	if _, err := c.fs.Stat(cfgFilePath); err == nil {
		// configuration file already exists, then do nothing.
		logE.WithFields(logrus.Fields{
			"configuration_file_path": cfgFilePath,
		}).Info("configuration file already exists")
		return nil
	}

	registryVersion := "v4.48.2" // renovate: depName=aquaproj/aqua-registry
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
	cfgStr := strings.Replace(configTemplate, "%%STANDARD_REGISTRY_VERSION%%", registryVersion, 1)
	if err := afero.WriteFile(c.fs, cfgFilePath, []byte(cfgStr), osfile.FilePermission); err != nil {
		return fmt.Errorf("write a configuration file: %w", logerr.WithFields(err, logrus.Fields{
			"configuration_file_path": cfgFilePath,
		}))
	}
	return nil
}
