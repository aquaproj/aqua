package rmtoken

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func (c *Controller) Remove(logE *logrus.Entry) error {
	if err := c.tokenManager.Remove(logE); err != nil {
		return fmt.Errorf("remove a GitHub access Token from the secret store: %w", err)
	}
	return nil
}
