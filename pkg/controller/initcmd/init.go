package initcmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	githubSvc "github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
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

var globalVarCfgFileNameMap = map[string]struct{}{ //nolint:gochecknoglobals
	"aqua.yaml":  {},
	"aqua.yml":   {},
	".aqua.yaml": {},
	".aqua.yml":  {},
}

type Controller struct {
	gitHubRepositoryService githubSvc.RepositoryService
}

func New(gh githubSvc.RepositoryService) *Controller {
	return &Controller{
		gitHubRepositoryService: gh,
	}
}

func (ctrl *Controller) Init(ctx context.Context, cfgFilePath string, logE *logrus.Entry) error {
	if cfgFilePath == "" {
		cfgFilePath = "aqua.yaml"
	}
	if _, ok := globalVarCfgFileNameMap[cfgFilePath]; ok {
		for fileName := range globalVarCfgFileNameMap {
			if _, err := os.Stat(fileName); err == nil {
				// configuration file already exists, then do nothing.
				logE.WithFields(logrus.Fields{
					"configuration_file_path": fileName,
				}).Info("configuration file already exists")
				return nil
			}
		}
	}
	if _, err := os.Stat(cfgFilePath); err == nil {
		// configuration file already exists, then do nothing.
		return nil
	}
	registryVersion := "v2.11.0" // renovate: depName=aquaproj/aqua-registry
	release, _, err := ctrl.gitHubRepositoryService.GetLatestRelease(ctx, "aquaproj", "aqua-registry")
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
	if err := os.WriteFile(cfgFilePath, []byte(cfgStr), 0o644); err != nil { //nolint:gosec,gomnd
		return fmt.Errorf("write a configuration file: %w", logerr.WithFields(err, logrus.Fields{
			"configuration_file_path": cfgFilePath,
		}))
	}
	return nil
}
