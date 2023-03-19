package initpolicy

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const configTemplate = `---
# aqua Policy
# https://aquaproj.github.io/
registries:
- type: standard
  ref: semver(">= 3.0.0")
packages:
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

func (ctrl *Controller) Init(ctx context.Context, cfgFilePath string, logE *logrus.Entry) error {
	if cfgFilePath == "" {
		cfgFilePath = "aqua-policy.yaml"
	}
	if _, err := ctrl.fs.Stat(cfgFilePath); err == nil {
		// configuration file already exists, then do nothing.
		logE.WithFields(logrus.Fields{
			"policy_file_path": cfgFilePath,
		}).Info("plicy file already exists")
		return nil
	}

	if err := afero.WriteFile(ctrl.fs, cfgFilePath, []byte(configTemplate), 0o644); err != nil { //nolint:gomnd
		return fmt.Errorf("write a policy file: %w", logerr.WithFields(err, logrus.Fields{
			"policy_file_path": cfgFilePath,
		}))
	}
	return nil
}
