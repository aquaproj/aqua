package vacuum

import (
	"context"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
	"github.com/sirupsen/logrus"
)

func (c *Controller) Vacuum(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	timestamps, err := c.vacuum.FindAll()
	if err != nil {
		return err
	}
	timestampChecker := vacuum.NewTimestampChecker(time.Now(), param.VacuumDays)
	for pkgPath, timestamp := range timestamps {
		logE := logE.WithField("package_path", pkgPath)
		if !timestampChecker.Expired(timestamp) {
			continue
		}
		// remove the package
		p := filepath.Join(c.rootDir, pkgPath)
		if err := c.fs.RemoveAll(p); err != nil {
			return err
		}
		// remove the timestamp file
		if err := c.vacuum.Remove(pkgPath); err != nil {
			return err
		}
		logE.Info("removed the package")
	}
	return nil
}
