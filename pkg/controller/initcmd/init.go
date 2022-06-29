package initcmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const configTemplate = `---
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
registries:
- type: standard
  ref: %%STANDARD_REGISTRY_VERSION%% # renovate: depName=aquaproj/aqua-registry
packages:
`

type Controller struct {
	github RepositoriesService
	fs     afero.Fs
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
}

func New(gh RepositoriesService, fs afero.Fs) *Controller {
	return &Controller{
		github: gh,
		fs:     fs,
	}
}

func (ctrl *Controller) Init(ctx context.Context, cfgFilePath string, logE *logrus.Entry) error {
	if cfgFilePath == "" {
		cfgFilePath = "aqua.yaml"
	}
	if _, err := ctrl.fs.Stat(cfgFilePath); err == nil {
		// configuration file already exists, then do nothing.
		logE.WithFields(logrus.Fields{
			"configuration_file_path": cfgFilePath,
		}).Info("configuration file already exists")
		return nil
	}

	registryVersion := "v3.4.0" // renovate: depName=aquaproj/aqua-registry
	release, _, err := ctrl.github.GetLatestRelease(ctx, "aquaproj", "aqua-registry")
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
	if err := afero.WriteFile(ctrl.fs, cfgFilePath, []byte(cfgStr), 0o644); err != nil { //nolint:gomnd
		return fmt.Errorf("write a configuration file: %w", logerr.WithFields(err, logrus.Fields{
			"configuration_file_path": cfgFilePath,
		}))
	}
	return nil
}
