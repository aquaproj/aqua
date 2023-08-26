package initpolicy

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const configTemplate = `---
# aqua Policy
# https://aquaproj.github.io/
registries:
# Example
# - name: local
#   type: local
#   path: registry.yaml
# - name: aqua-registry
#   type: github_content
#   repo_owner: aquaproj
#   repo_name: aqua-registry
#   ref: semver(">= 3.0.0") # ref is optional
#   path: registry.yaml
  - type: standard
    ref: semver(">= 3.0.0")
packages:
# Example
# - registry: local # allow all packages in the Registry
# - name: cli/cli # allow only a specific package. The default value of registry is "standard"
# - name: cli/cli
#   version: semver(">= 2.0.0") # version is optional
  - registry: standard
`

type Controller struct {
	fs afero.Fs
}

func New(fs afero.Fs) *Controller {
	return &Controller{
		fs: fs,
	}
}

func (c *Controller) Init(ctx context.Context, cfgFilePath string, logE *logrus.Entry) error {
	if cfgFilePath == "" {
		cfgFilePath = "aqua-policy.yaml"
	}
	if _, err := c.fs.Stat(cfgFilePath); err == nil {
		// configuration file already exists, then do nothing.
		logE.WithFields(logrus.Fields{
			"policy_file_path": cfgFilePath,
		}).Info("plicy file already exists")
		return nil
	}

	if err := afero.WriteFile(c.fs, cfgFilePath, []byte(configTemplate), osfile.FilePermission); err != nil {
		return fmt.Errorf("write a policy file: %w", logerr.WithFields(err, logrus.Fields{
			"policy_file_path": cfgFilePath,
		}))
	}
	return nil
}
