package disallowpolicy

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	fs                 afero.Fs
	policyConfigFinder policy.ConfigFinder
	policyValidator    policy.Validator
}

func New(fs afero.Fs, policyConfigFinder policy.ConfigFinder, policyValidator policy.Validator) *Controller {
	return &Controller{
		fs:                 fs,
		policyConfigFinder: policyConfigFinder,
		policyValidator:    policyValidator,
	}
}

func (ctrl *Controller) Disallow(ctx context.Context, logE *logrus.Entry, param *config.Param, policyFilePath string) error {
	policyFilePath, err := ctrl.policyConfigFinder.Find(policyFilePath, param.PWD)
	if err != nil {
		return fmt.Errorf("find a policy file: %w", err)
	}
	if policyFilePath == "" {
		logE.Info("no policy file is found")
		return nil
	}
	if err := ctrl.policyValidator.Disallow(policyFilePath); err != nil {
		return logerr.WithFields(fmt.Errorf("disallow a policy file: %w", err), logrus.Fields{ //nolint:wrapcheck
			"policy_file": policyFilePath,
		})
	}
	return nil
}
