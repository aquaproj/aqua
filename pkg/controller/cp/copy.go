package cp

import (
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/sirupsen/logrus"
)

func (c *Controller) copy(logE *logrus.Entry, param *config.Param, findResult *which.FindResult, exeName string) error {
	p := filepath.Join(param.Dest, exeName)
	if c.runtime.GOOS == "windows" && filepath.Ext(exeName) == "" {
		p += ".exe"
	}
	logE.WithFields(logrus.Fields{
		"exe_name": exeName,
		"dest":     p,
	}).Info("coping a file")
	if err := c.packageInstaller.Copy(p, findResult.ExePath); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}
