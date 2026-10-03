package updatepolicy

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
)

func (c *Controller) UpdatePolicy(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	// Find a policy file
	// Find aqua.yaml
	// list packages
	// Update a policy file
	var policyCfgs []*policy.Config
	for _, cfgFilePath := range c.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		policyCfgs, err := c.policyReader.Append(logE, cfgFilePath, policyCfgs, nil)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := c.install(ctx, logE, cfgFilePath, policyCfgs, param); err != nil {
			return err
		}
	}
	return nil
}
