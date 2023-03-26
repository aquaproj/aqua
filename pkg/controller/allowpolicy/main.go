package allowpolicy

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/policy"
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

func (ctrl *Controller) Allow(ctx context.Context, logE *logrus.Entry, param *config.Param, policyFilePath string) error {
	policyFile, err := ctrl.policyConfigFinder.Find(policyFilePath, param.PWD)
	if err != nil {
		return fmt.Errorf("find a policy file: %w", err)
	}
	if policyFile == "" {
		logE.Info("no policy file is found")
		return nil
	}
	if err := ctrl.policyValidator.Allow(policyFile); err != nil {
		return logerr.WithFields(fmt.Errorf("allow a policy file: %w", err), logrus.Fields{ //nolint:wrapcheck
			"policy_file": policyFile,
		})
	}
	return nil
}
